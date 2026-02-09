# Temporal Workflow Engine Research for ERP Workflow Automation

**Date:** February 2026
**Status:** Research Complete

## Executive Summary

**Your pattern is valid and well-supported by Temporal.** The "generic interpreter workflow" that executes arbitrary graph structures passed as input is a recognized pattern with community examples and a dedicated library (temporalgraph). However, there are important architectural considerations and trade-offs to understand.

---

## 1. Is a Generic Interpreter Workflow a Valid Temporal Pattern?

**Yes, this is a valid and documented pattern.**

### Evidence from Temporal Community

- Multiple community discussions confirm teams are building DSL-based DAG workflows defined from UI
- The DSL Workflow sample in temporalio/samples-go demonstrates executing workflows from YAML configuration files, enabling a "low code" layer
- The Dynamic Workflows sample shows executing workflows dynamically using a single "Dynamic Workflow"

### Determinism Considerations

**Key insight:** Your workflow logic can depend on input data without breaking determinism, as long as the execution produces the same sequence of Commands given the same History.

From Temporal documentation:
> "For a given Workflow Type, its Workflow Definition (implementation) must produce the same sequence of Commands given the same History."

**What this means for your use case:**
- The graph structure passed as input is part of the workflow input, which is stored in history
- On replay, the same graph structure is available, so your interpreter will produce the same command sequence
- **Safe:** Branching based on graph edges, node types, edge conditions
- **Unsafe:** Calling `time.Now()`, `rand.Int()`, or making external API calls directly in workflow code

**Best Practice:** All non-deterministic operations (external calls, random values, current time) must happen inside Activities, not in the workflow interpreter logic.

---

## 2. Parallel Execution with Convergence

### Can workflow.Go() goroutines share and mutate a context struct safely?

**Yes, with important caveats.** From the Go SDK Multithreading docs:

> "Temporal's Go SDKs contains a deterministic runner to control the thread execution. This deterministic runner will decide which Workflow thread to run in the right order, and one at a time."

This means:
- Temporal goroutines are **never truly concurrent** - they're cooperatively scheduled
- Race conditions are minimized, but not impossible during yielding points
- **Recommendation:** Use workflow channels for communication between goroutines rather than shared mutable state

### workflow.Go() vs Child Workflows

| Pattern | Use When |
|---------|----------|
| `workflow.Go()` | Parallel branches within same execution context, shared state needed |
| Child Workflows | Long-running branches that need their own history, isolation, or Continue-As-New |

For your use case (parallel action execution with merged context), `workflow.Go()` is the right choice.

### Waiting for Parallel Paths to Complete (Convergence/Join)

The Split/Merge Future sample demonstrates the pattern:

```go
// Launch parallel activities
var futures []workflow.Future
for _, branch := range parallelBranches {
    f := workflow.ExecuteActivity(ctx, ExecuteAction, branch)
    futures = append(futures, f)
}

// Wait for all to complete (convergence)
for _, f := range futures {
    var result ActionResult
    if err := f.Get(ctx, &result); err != nil {
        return err
    }
    // Merge result into context
    executionContext.Merge(result)
}
```

Alternative using `workflow.Selector` for processing results as they arrive:

```go
selector := workflow.NewSelector(ctx)
for _, f := range futures {
    selector.AddFuture(f, func(f workflow.Future) {
        var result ActionResult
        f.Get(ctx, &result)
        executionContext.Merge(result)
    })
}
// Wait for all
for range futures {
    selector.Select(ctx)
}
```

---

## 3. Long-Running Async Operations (RabbitMQ Integration)

### Pattern: Asynchronous Activity Completion

For actions that publish to RabbitMQ and wait for a response, use Asynchronous Activity Completion:

```go
// In your Activity
func PublishAndWaitForResponse(ctx context.Context, message Message) (string, error) {
    // Get task token for async completion
    activityInfo := activity.GetInfo(ctx)
    taskToken := activityInfo.TaskToken

    // Publish message with task token as correlation ID
    rabbitMQ.Publish(message, taskToken)

    // Return pending - activity doesn't complete yet
    return "", activity.ErrResultPending
}
```

```go
// In your RabbitMQ consumer (separate service)
func OnMessageReceived(response Response) {
    temporalClient.CompleteActivity(ctx, response.TaskToken, response.Result, nil)
}
```

### When to Use Signals vs Async Activity Completion

| Approach | Use When |
|----------|----------|
| **Async Activity Completion** | External system might fail, need heartbeats, need cancellation propagation |
| **Signals** | External system is reliable, simpler implementation, don't need activity retry semantics |

For RabbitMQ with guaranteed delivery, either works. Async Activity Completion is slightly more robust.

---

## 4. Human-in-the-Loop Workflows

### Recommended Pattern: Signals with Wait Condition

```go
// Workflow code
var approved *bool

// Define signal handler
workflow.SetSignalChannel(ctx, "approval", approvalChannel)

// Wait for approval (can wait days/weeks)
workflow.Go(ctx, func(ctx workflow.Context) {
    var approval ApprovalSignal
    approvalChannel.Receive(ctx, &approval)
    approved = &approval.Approved
})

// Wait with optional timeout
ok, _ := workflow.AwaitWithTimeout(ctx, 5*24*time.Hour, func() bool {
    return approved != nil
})

if !ok {
    // Timeout - handle escalation
}
```

**Key benefit:** While waiting, the workflow consumes **zero compute resources**. The worker isn't polling - Temporal's server handles the timer and signal routing.

> "Temporal handles timers and signals so you can easily do things like sleep for 30 days or wait for a human approval without holding a thread."

---

## 5. Context/State Accumulation

### Is passing a growing context struct idiomatic?

**Yes, but with size limits to consider.**

| Limit | Threshold |
|-------|-----------|
| Single payload | 2 MB hard limit (warning at 256 KB) |
| gRPC message | 4 MB |
| Event history | 50 MB or 51,200 events |

### Performance Considerations

> "The entire Execution history is transferred from the Temporal Service to Workflow Workers when a Workflow state needs to recover. A large Execution history can thus adversely impact the performance of your Workflow."

### Best Practices for Large State

1. **Keep context lean:** Store only what's needed for subsequent actions
2. **Offload large data:** Store large payloads in external storage (S3, database), pass references
3. **Use Continue-As-New:** For very long-running workflows, periodically reset history while preserving state:

```go
if workflow.GetInfo(ctx).GetCurrentHistoryLength() > 10000 {
    return workflow.NewContinueAsNewError(ctx, WorkflowFunc, currentState)
}
```

---

## 6. Production Infrastructure Requirements

### Database Options

Temporal supports three persistence backends:

| Database | Production Ready | Notes |
|----------|------------------|-------|
| PostgreSQL | Yes | Fully supported, most common for new deployments |
| MySQL | Yes | Fully supported |
| Cassandra | Yes | Best for very high scale |

**PostgreSQL works perfectly for your use case.** PostgreSQL is commonly used in production Temporal deployments.

### Self-Hosted Architecture Components

A production Temporal deployment requires:

1. **Database:** PostgreSQL (you already have this)
2. **Temporal Server** (4 services):
   - Frontend
   - Matching
   - History
   - Worker (internal)
3. **Elasticsearch** (optional): Required only for advanced visibility features (complex workflow queries)
4. **Web UI** (optional): For debugging and operations

> "Self hosting is daunting. Temporal is not a single process with a single run command. It's a persistent database and a cluster of multiple processes."

### Deployment Options

1. **Docker Compose:** Quickest for development/testing
2. **Kubernetes via Helm:** Production-grade, uses temporalio/helm-charts
3. **Temporal Cloud:** Managed service, eliminates operational burden

### Operational Complexity Comparison

| Aspect | Self-Hosted | Temporal Cloud |
|--------|-------------|----------------|
| Initial Setup | High (days-weeks) | Low (hours) |
| Scaling | Manual DB + services | Automatic |
| Upgrades | Manual schema migrations | Automatic |
| Cost | Infrastructure + engineering time | Consumption-based |
| Performance | Varies by setup | ~2x faster (benchmark) |

---

## Red Flags and Significant Challenges

### 1. Determinism Violations (Critical)

Your interpreter **must not**:
- Use `time.Now()` or `rand` in workflow code
- Make HTTP calls directly
- Access databases directly
- Use non-deterministic iteration (e.g., map iteration order)

**Solution:** All external operations go in Activities.

### 2. Event History Growth

With complex workflows, history can grow large. Monitor and implement Continue-As-New for workflows with many actions.

### 3. Workflow Versioning

If you change your interpreter logic while workflows are running, you need versioning:

```go
v := workflow.GetVersion(ctx, "interpreter-v2", workflow.DefaultVersion, 1)
if v == workflow.DefaultVersion {
    // Old logic
} else {
    // New logic
}
```

### 4. Testing Complexity

Testing dynamic workflows requires:
- Replay testing with historical events
- Testing various graph topologies
- Edge case handling (cycles, orphan nodes)

### 5. Operational Learning Curve

Self-hosting Temporal requires understanding:
- Task queues and worker deployment
- Namespace management
- Visibility and debugging tools
- Database scaling

---

## Alternative Patterns to Consider

### 1. Temporalgraph Library

The temporalgraph library provides:
- DAG composition with AddNode, AddEdge, Compile, Invoke API
- Type-safe node inputs/outputs
- Fan-out/fan-in as first-class patterns
- Conditional routing

This could accelerate your implementation.

### 2. Compile-Time Workflow Generation

Instead of runtime interpretation, generate Go workflow code from your graph definition:
- Pros: Better type safety, easier debugging, no interpreter complexity
- Cons: Requires code deployment for workflow changes

### 3. Hybrid Approach

Use Temporal for durable execution with a simpler activity-per-action model:
- Workflow receives graph, does topological sort
- Iterates through nodes, calling a generic `ExecuteAction` activity
- Activity looks up action type and executes appropriately

This is cleaner than full interpreter but still dynamic.

---

## Recommended Architecture

Based on your requirements, here's a suggested approach:

```
┌─────────────────────────────────────────────────────────────┐
│  Workflow: GraphExecutor                                     │
│  - Input: GraphDefinition (nodes, edges, start conditions)  │
│  - State: ExecutionContext (accumulates action results)     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Logic: Traverse graph from start edges                      │
│  1. Find executable nodes (all incoming edges satisfied)    │
│  2. For parallel edges: workflow.Go() for each branch       │
│  3. Execute node via Activity                               │
│  4. Merge result into ExecutionContext                      │
│  5. Evaluate outgoing edges (conditions, branching)         │
│  6. Repeat until terminal node                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Activities:                                                 │
│  - ExecuteAction(nodeID, actionType, context) → result      │
│  - SendRabbitMQ(message) → async completion                 │
│  - SendEmail(template, context)                             │
│  - UpdateRecord(entity, changes)                            │
│  - WaitForApproval() → signal-based                         │
└─────────────────────────────────────────────────────────────┘
```

---

## Sources

### Official Documentation
- [Temporal Workflow Documentation](https://docs.temporal.io/workflows)
- [Go SDK Multithreading](https://docs.temporal.io/develop/go/go-sdk-multithreading)
- [Asynchronous Activity Completion - Go SDK](https://docs.temporal.io/develop/go/asynchronous-activity-completion)
- [Self-hosted Temporal Service defaults](https://docs.temporal.io/self-hosted-guide/defaults)
- [Deploying a Temporal Service](https://docs.temporal.io/self-hosted-guide/deployment)
- [Workflow Execution limits](https://docs.temporal.io/workflow-execution/limits)

### Samples and Libraries
- [Temporal Go SDK Samples](https://github.com/temporalio/samples-go)
- [Temporalgraph - Graph-based Orchestration](https://temporal.io/code-exchange/temporalgraph-graph-based-orchestration)
- [Temporal Helm Charts](https://github.com/temporalio/helm-charts)

### Community Discussions
- [Executing a DAG in a workflow](https://community.temporal.io/t/executing-a-dag-in-a-workflow/8472)
- [Workflow for running a DAG in DSL](https://community.temporal.io/t/workflow-for-running-a-dag-in-dsl/3880)
- [Human dependent long running Workflows](https://community.temporal.io/t/human-dependent-long-running-workflows/3403)
- [PostgreSQL: good option for persistence in production?](https://community.temporal.io/t/postgresql-good-option-for-persistence-in-production/6153)

### Blog Posts and Tutorials
- [Managing very long-running Workflows with Temporal](https://temporal.io/blog/very-long-running-workflows)
- [Adding Durable Human-in-the-Loop](https://learn.temporal.io/tutorials/ai/building-durable-ai-applications/human-in-the-loop/)
- [Cloud Benchmark: Temporal Cloud vs. Self-Hosted](https://temporal.io/blog/benchmarking-latency-temporal-cloud-vs-self-hosted-temporal)
- [Reliable data processing: Queues and Workflows](https://temporal.io/blog/reliable-data-processing-queues-workflows)

---

## Conclusion

Temporal is a strong fit for your workflow automation engine. The generic interpreter pattern is valid, PostgreSQL is fully supported for production, and all your requirements (parallel execution, async continuation, human-in-the-loop, merged context) have well-documented solutions.

**Key decision point:** Self-hosted vs Temporal Cloud. Self-hosted gives you full control but requires significant operational investment. Temporal Cloud eliminates that burden but adds a dependency on their service.

**Recommended next step:** Build a proof-of-concept with a simple 3-node workflow (start → async action → end) to validate the integration pattern with your existing codebase.
