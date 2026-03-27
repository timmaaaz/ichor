# Receiving Discrepancy Notes Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> **Worktree:** Create a worktree before executing: `create a worktree for blocker-011-receiving-notes and execute this plan`

**Goal:** Add Notes field to ReceiveQuantityRequest so floor workers' discrepancy notes are persisted during receiving.

**Architecture:** Thread a `notes` parameter through 4 layers (API handler -> app model -> app method -> bus method). No migration needed -- DB column and bus model already support it.

**Tech Stack:** Go 1.23, Ardan Labs service architecture

---

## Step 1: Add Notes field to ReceiveQuantityRequest (app model)

- [ ] Edit `app/domain/procurement/purchaseorderlineitemapp/model.go`

At line 344-347, add the `Notes` field to the struct:

```go
// ReceiveQuantityRequest represents a request to receive quantity for a line item.
type ReceiveQuantityRequest struct {
	Quantity   string  `json:"quantity" validate:"required"`
	ReceivedBy string  `json:"received_by" validate:"required"`
	Notes      *string `json:"notes"`
}
```

`Notes` is `*string` (pointer) because it is optional -- nil means "don't change notes", empty string means "clear notes".

**Verify:** `go build ./app/domain/procurement/purchaseorderlineitemapp/...`

---

## Step 2: Add notes parameter to business method

- [ ] Edit `business/domain/procurement/purchaseorderlineitembus/purchaseorderlineitembus.go`

At line 245, update the `ReceiveQuantity` signature to accept an optional notes parameter:

```go
// ReceiveQuantity updates the received quantity for a line item.
func (b *Business) ReceiveQuantity(ctx context.Context, poli PurchaseOrderLineItem, quantity int, receivedBy uuid.UUID, notes *string) (PurchaseOrderLineItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitembus.receivequantity")
	defer span.End()

	before := poli

	poli.QuantityReceived += quantity
	poli.UpdatedBy = receivedBy
	poli.UpdatedDate = time.Now().UTC()

	if notes != nil {
		poli.Notes = *notes
	}

	if err := b.storer.Update(ctx, poli); err != nil {
		return PurchaseOrderLineItem{}, fmt.Errorf("receivequantity: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.del.Call(ctx, ActionUpdatedData(before, poli)); err != nil {
		b.log.Error(ctx, "purchaseorderlineitembus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return poli, nil
}
```

**Verify:** `go build ./business/domain/procurement/purchaseorderlineitembus/...`

---

## Step 3: Thread notes through the app layer

- [ ] Edit `app/domain/procurement/purchaseorderlineitemapp/purchaseorderlineitemapp.go`

At line 167, update the `ReceiveQuantity` method to accept and forward notes:

```go
// ReceiveQuantity updates the received quantity for a line item.
func (a *App) ReceiveQuantity(ctx context.Context, id uuid.UUID, quantity int, receivedBy uuid.UUID, notes *string) (PurchaseOrderLineItem, error) {
	poli, err := a.purchaseorderlineitembus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderlineitembus.ErrNotFound) {
			return PurchaseOrderLineItem{}, errs.New(errs.NotFound, err)
		}
		return PurchaseOrderLineItem{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	updatedPOLI, err := a.purchaseorderlineitembus.ReceiveQuantity(ctx, poli, quantity, receivedBy, notes)
	if err != nil {
		return PurchaseOrderLineItem{}, errs.Newf(errs.Internal, "receivequantity: purchaseorderlineitem[%+v]: %s", updatedPOLI, err)
	}

	return ToAppPurchaseOrderLineItem(updatedPOLI), nil
}
```

**Verify:** `go build ./app/domain/procurement/purchaseorderlineitemapp/...`

---

## Step 4: Extract and pass notes in the API handler

- [ ] Edit `api/domain/http/procurement/purchaseorderlineitemapi/purchaseorderlineitemapi.go`

At line 134, update the `receiveQuantity` handler to extract `app.Notes` and pass it through:

```go
func (api *api) receiveQuantity(ctx context.Context, r *http.Request) web.Encoder {
	var app purchaseorderlineitemapp.ReceiveQuantityRequest
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	poliID := web.Param(r, "purchase_order_line_item_id")
	parsed, err := uuid.Parse(poliID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	quantity, err := strconv.Atoi(app.Quantity)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	receivedBy, err := uuid.Parse(app.ReceivedBy)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	poli, err := api.purchaseorderlineitemapp.ReceiveQuantity(ctx, parsed, quantity, receivedBy, app.Notes)
	if err != nil {
		return errs.NewError(err)
	}

	return poli
}
```

**Verify:** `go build ./api/domain/http/procurement/purchaseorderlineitemapi/...`

---

## Step 5: Fix any other callers of ReceiveQuantity

- [ ] Search for all call sites of `ReceiveQuantity` across the codebase and add the `notes` parameter (pass `nil` where notes are not applicable).

Likely callers to check:
- Integration tests in `api/cmd/services/ichor/tests/procurement/`
- Any workflow action handlers that call ReceiveQuantity
- Any formdata processors

**Verify:** `go build ./...` (full build to catch all broken call sites)

---

## Step 6: Run tests

- [ ] Run tests for all affected packages:

```bash
go test ./business/domain/procurement/purchaseorderlineitembus/... \
       ./app/domain/procurement/purchaseorderlineitemapp/... \
       ./api/domain/http/procurement/purchaseorderlineitemapi/...
```

Expected: all tests pass.

---

## Step 7: Commit

- [ ] Commit with message: `feat(procurement): add notes field to ReceiveQuantityRequest for discrepancy tracking`
