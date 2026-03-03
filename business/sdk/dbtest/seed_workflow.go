package dbtest

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

func seedWorkflow(ctx context.Context, log *logger.Logger, busDomain BusDomain, adminID uuid.UUID) error {
	// =============================================================================
	// WORKFLOW AUTOMATION RULES FOR ORDER PROCESSING
	// =============================================================================

	log.Info(ctx, "🔄 Seeding workflow automation rules for order processing...")

	// First, ensure allocation_results entity exists for downstream workflow triggers
	// This is a virtual entity used by the allocation system to fire workflow events
	wfEntityType, err := busDomain.Workflow.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		log.Error(ctx, "Failed to query entity type 'table' for automation rules", "error", err)
		// Don't fail the entire seed - automation rules are enhancement, not critical
	} else {
		// Check if allocation_results entity exists, create if not
		_, err := busDomain.Workflow.QueryEntityByName(ctx, "allocation_results")
		if err != nil {
			// Create the virtual entity for allocation results
			_, createErr := busDomain.Workflow.CreateEntity(ctx, workflow.NewEntity{
				Name:         "allocation_results",
				EntityTypeID: wfEntityType.ID,
				SchemaName:   "workflow",
			})
			if createErr != nil {
				log.Error(ctx, "Failed to create allocation_results entity", "error", createErr)
			} else {
				log.Info(ctx, "✅ Created allocation_results entity for workflow events")
			}
		}

		// Query required entities and trigger types
		orderLineItemsEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "order_line_items")
		if err != nil {
			log.Error(ctx, "Failed to query order_line_items entity", "error", err)
		}

		allocationResultsEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "allocation_results")
		if err != nil {
			log.Error(ctx, "Failed to query allocation_results entity", "error", err)
		}

		onCreateTrigger, err := busDomain.Workflow.QueryTriggerTypeByName(ctx, "on_create")
		if err != nil {
			log.Error(ctx, "Failed to query on_create trigger type", "error", err)
		}

		// Create action templates for workflow actions
		allocateTemplate, err := busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Allocate Inventory",
			Description:   "Allocates inventory for an order or request",
			ActionType:    "allocate_inventory",
			Icon:          "material-symbols:inventory-2",
			DefaultConfig: json.RawMessage(`{}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create allocate_inventory template", "error", err)
		}

		updateFieldTemplate, err := busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Update Field",
			Description:   "Updates a field on the target entity",
			ActionType:    "update_field",
			Icon:          "material-symbols:edit-note",
			DefaultConfig: json.RawMessage(`{}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create update_field template", "error", err)
		}

		createAlertTemplate, err := busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Create Alert",
			Description:   "Creates an alert notification",
			ActionType:    "create_alert",
			Icon:          "material-symbols:notification-important",
			DefaultConfig: json.RawMessage(`{}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create create_alert template", "error", err)
		}

		// Granular inventory action templates
		checkInventoryTemplate, err := busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Check Inventory",
			Description:   "Checks inventory availability against a threshold",
			ActionType:    "check_inventory",
			Icon:          "material-symbols:fact-check",
			DefaultConfig: json.RawMessage(`{"threshold": 1}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create check_inventory template", "error", err)
		}

		reserveInventoryTemplate, err := busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Reserve Inventory",
			Description:   "Reserves inventory with idempotency support",
			ActionType:    "reserve_inventory",
			Icon:          "material-symbols:bookmark-added",
			DefaultConfig: json.RawMessage(`{"allocation_strategy":"fifo","reservation_duration_hours":24,"allow_partial":false}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create reserve_inventory template", "error", err)
		}

		_, err = busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Release Reservation",
			Description:   "Releases reserved inventory back to available stock",
			ActionType:    "release_reservation",
			Icon:          "material-symbols:remove-shopping-cart",
			DefaultConfig: json.RawMessage(`{}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create release_reservation template", "error", err)
		}

		_, err = busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Commit Allocation",
			Description:   "Commits reserved inventory to allocated status",
			ActionType:    "commit_allocation",
			Icon:          "material-symbols:check-circle",
			DefaultConfig: json.RawMessage(`{}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create commit_allocation template", "error", err)
		}

		checkReorderPointTemplate, err := busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Check Reorder Point",
			Description:   "Checks if inventory is below reorder point",
			ActionType:    "check_reorder_point",
			Icon:          "material-symbols:trending-down",
			DefaultConfig: json.RawMessage(`{}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create check_reorder_point template", "error", err)
		}

		logAuditEntryTemplate, err := busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Log Audit Entry",
			Description:   "Write an audit trail entry to the workflow audit log",
			ActionType:    "log_audit_entry",
			Icon:          "material-symbols:history-edu",
			DefaultConfig: json.RawMessage(`{}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create log_audit_entry template", "error", err)
		}

		_, err = busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Create Entity",
			Description:   "Create a new entity record in the database",
			ActionType:    "create_entity",
			Icon:          "material-symbols:note-add",
			DefaultConfig: json.RawMessage(`{}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create create_entity template", "error", err)
		}

		_, err = busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Delay",
			Description:   "Pause workflow execution for a specified duration",
			ActionType:    "delay",
			Icon:          "material-symbols:timer",
			DefaultConfig: json.RawMessage(`{"duration": "5m"}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create delay template", "error", err)
		}

		_, err = busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Evaluate Condition",
			Description:   "Evaluates conditions and determines branch direction",
			ActionType:    "evaluate_condition",
			Icon:          "material-symbols:fork-right",
			DefaultConfig: json.RawMessage(`{}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create evaluate_condition template", "error", err)
		}

		_, err = busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Lookup Entity",
			Description:   "Look up entity data by filter criteria",
			ActionType:    "lookup_entity",
			Icon:          "material-symbols:manage-search",
			DefaultConfig: json.RawMessage(`{}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create lookup_entity template", "error", err)
		}

		_, err = busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Seek Approval",
			Description:   "Creates an approval request for specified users",
			ActionType:    "seek_approval",
			Icon:          "material-symbols:approval",
			DefaultConfig: json.RawMessage(`{}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create seek_approval template", "error", err)
		}

		_, err = busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Send Email",
			Description:   "Sends an email to specified recipients",
			ActionType:    "send_email",
			Icon:          "material-symbols:forward-to-inbox",
			DefaultConfig: json.RawMessage(`{}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create send_email template", "error", err)
		}

		sendNotificationTemplate, err := busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Send Notification",
			Description:   "Sends in-app notifications through various channels",
			ActionType:    "send_notification",
			Icon:          "material-symbols:campaign",
			DefaultConfig: json.RawMessage(`{}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create send_notification template", "error", err)
		}

		_, err = busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Transition Status",
			Description:   "Transition an entity field from one status to another",
			ActionType:    "transition_status",
			Icon:          "material-symbols:swap-horiz",
			DefaultConfig: json.RawMessage(`{}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create transition_status template", "error", err)
		}

		_, err = busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Call Webhook",
			Description:   "Makes an outbound HTTP request to an external URL",
			ActionType:    "call_webhook",
			Icon:          "material-symbols:webhook",
			DefaultConfig: json.RawMessage(`{"method": "POST", "timeout_seconds": 30}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create call_webhook template", "error", err)
		}

		_, err = busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Receive Inventory",
			Description:   "Receives inventory into a warehouse location from a purchase order",
			ActionType:    "receive_inventory",
			Icon:          "material-symbols:local-shipping",
			DefaultConfig: json.RawMessage(`{"source_from_po": true}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create receive_inventory template", "error", err)
		}

		_, err = busDomain.Workflow.CreateActionTemplate(ctx, workflow.NewActionTemplate{
			Name:          "Create Purchase Order",
			Description:   "Creates a purchase order with line items for supplier procurement",
			ActionType:    "create_purchase_order",
			Icon:          "material-symbols:receipt-long",
			DefaultConfig: json.RawMessage(`{}`),
			CreatedBy:     adminID,
		})
		if err != nil {
			log.Error(ctx, "Failed to create create_purchase_order template", "error", err)
		}

		// Create automation rules if we have all the required references
		if orderLineItemsEntity.ID != uuid.Nil && wfEntityType.ID != uuid.Nil && onCreateTrigger.ID != uuid.Nil {
			// Rule 1: Line Item Created -> Allocate Inventory
			// Triggers on each order_line_items.on_create, extracts product_id/quantity from RawData
			allocateConfig := map[string]interface{}{
				"source_from_line_item": true, // Extract product_id, quantity, order_id from line item RawData
				"allocation_mode":       "reserve",
				"allocation_strategy":   "fifo",
				"allow_partial":         false,
				"priority":              "high",
				"reference_type":        "order",
			}
			allocateConfigJSON, _ := json.Marshal(allocateConfig)

			orderAllocateRule, err := busDomain.Workflow.CreateRule(ctx, workflow.NewAutomationRule{
				Name:          "Line Item Created - Allocate Inventory",
				Description:   "When an order line item is created, attempt to reserve inventory for that product",
				EntityID:      orderLineItemsEntity.ID,
				EntityTypeID:  wfEntityType.ID,
				TriggerTypeID: onCreateTrigger.ID,
				IsActive:      false,
				IsDefault:     true,
				CreatedBy:     adminID,
			})
			if err != nil {
				log.Error(ctx, "Failed to create Line Item Allocate rule", "error", err)
			} else {
				allocateAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: orderAllocateRule.ID,
					Name:             "Allocate Inventory for Line Item",
					Description:      "Reserve inventory for the line item's product",
					ActionConfig:     json.RawMessage(allocateConfigJSON),
					IsActive:         true,
					TemplateID:       &allocateTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create allocate inventory action", "error", err)
				} else {
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         orderAllocateRule.ID,
						SourceActionID: nil,
						TargetActionID: allocateAction.ID,
						EdgeType:       "start",
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create edge for allocate inventory action", "error", err)
					}
					log.Info(ctx, "✅ Created 'Line Item Created - Allocate Inventory' rule")
				}
			}
		}

		if allocationResultsEntity.ID != uuid.Nil && wfEntityType.ID != uuid.Nil && onCreateTrigger.ID != uuid.Nil {
			// Rule 2: Allocation Success -> Update Line Item Status
			successCondition := map[string]interface{}{
				"field_conditions": []map[string]interface{}{
					{
						"field_name": "status",
						"operator":   "equals",
						"value":      "success",
					},
				},
			}
			successConditionJSON, _ := json.Marshal(successCondition)
			successConditionRaw := json.RawMessage(successConditionJSON)

			updateConfig := map[string]interface{}{
				"target_entity": "sales.order_line_items",
				"target_field":  "line_item_fulfillment_statuses_id",
				"new_value":     "ALLOCATED",
				"field_type":    "foreign_key",
				"foreign_key_config": map[string]interface{}{
					"reference_table": "sales.line_item_fulfillment_statuses",
					"lookup_field":    "name",
				},
				"conditions": []map[string]interface{}{
					{"field_name": "order_id", "operator": "equals", "value": "{{reference_id}}"},
				},
			}
			updateConfigJSON, _ := json.Marshal(updateConfig)

			allocationSuccessRule, err := busDomain.Workflow.CreateRule(ctx, workflow.NewAutomationRule{
				Name:              "Allocation Success - Update Line Items",
				Description:       "When inventory allocation succeeds, update order line items to ALLOCATED status",
				EntityID:          allocationResultsEntity.ID,
				EntityTypeID:      wfEntityType.ID,
				TriggerTypeID:     onCreateTrigger.ID,
				TriggerConditions: &successConditionRaw,
				IsActive:          false,
				IsDefault:     true,
				CreatedBy:         adminID,
			})
			if err != nil {
				log.Error(ctx, "Failed to create Allocation Success rule", "error", err)
			} else {
				updateAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: allocationSuccessRule.ID,
					Name:             "Update Line Items to ALLOCATED",
					Description:      "Set line item status to ALLOCATED after successful inventory reservation",
					ActionConfig:     json.RawMessage(updateConfigJSON),
					IsActive:         true,
					TemplateID:       &updateFieldTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create update line items action", "error", err)
				} else {
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         allocationSuccessRule.ID,
						SourceActionID: nil,
						TargetActionID: updateAction.ID,
						EdgeType:       "start",
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create edge for update line items action", "error", err)
					}

					// Add success alert action after the update
					successAlertConfig := map[string]interface{}{
						"alert_type": "inventory_allocation_success",
						"severity":   "critical",
						"title":      "Inventory Allocation Success",
						"message":    "success",
						"recipients": map[string]interface{}{
							"users": []string{"5cf37266-3473-4006-984f-9325122678b7"}, // Admin Gopher
							"roles": []string{},
						},
					}
					successAlertConfigJSON, _ := json.Marshal(successAlertConfig)

					successAlertAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
						AutomationRuleID: allocationSuccessRule.ID,
						Name:             "Alert Admin - Allocation Success",
						Description:      "Send critical alert to admin on successful allocation",
						ActionConfig:     json.RawMessage(successAlertConfigJSON),
						IsActive:         true,
						TemplateID:       &createAlertTemplate.ID,
					})
					if err != nil {
						log.Error(ctx, "Failed to create success alert action", "error", err)
					} else {
						// Chain: updateAction -> successAlertAction
						_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
							RuleID:         allocationSuccessRule.ID,
							SourceActionID: &updateAction.ID,
							TargetActionID: successAlertAction.ID,
							EdgeType:       "sequence",
							EdgeOrder:      1,
						})
						if err != nil {
							log.Error(ctx, "Failed to create edge for success alert action", "error", err)
						}
					}

					log.Info(ctx, "✅ Created 'Allocation Success - Update Line Items' rule with alert")
				}
			}

			// Rule 3: Allocation Failure -> Create Alert
			failedCondition := map[string]interface{}{
				"field_conditions": []map[string]interface{}{
					{
						"field_name": "status",
						"operator":   "equals",
						"value":      "failed",
					},
				},
			}
			failedConditionJSON, _ := json.Marshal(failedCondition)
			failedConditionRaw := json.RawMessage(failedConditionJSON)

			alertConfig := map[string]interface{}{
				"alert_type": "inventory_allocation_failed",
				"severity":   "critical",
				"title":      "Inventory Allocation Failed",
				"message":    "failed",
				"recipients": map[string]interface{}{
					"users": []string{"5cf37266-3473-4006-984f-9325122678b7"}, // Admin Gopher from seed.sql
					"roles": []string{},
				},
			}
			alertConfigJSON, _ := json.Marshal(alertConfig)

			allocationFailedRule, err := busDomain.Workflow.CreateRule(ctx, workflow.NewAutomationRule{
				Name:              "Allocation Failed - Alert Operations",
				Description:       "When inventory allocation fails, create an alert for the operations team",
				EntityID:          allocationResultsEntity.ID,
				EntityTypeID:      wfEntityType.ID,
				TriggerTypeID:     onCreateTrigger.ID,
				TriggerConditions: &failedConditionRaw,
				IsActive:          false,
				IsDefault:     true,
				CreatedBy:         adminID,
			})
			if err != nil {
				log.Error(ctx, "Failed to create Allocation Failed rule", "error", err)
			} else {
				alertAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: allocationFailedRule.ID,
					Name:             "Create Alert for Operations",
					Description:      "Notify operations team of allocation failure",
					ActionConfig:     json.RawMessage(alertConfigJSON),
					IsActive:         true,
					TemplateID:       &createAlertTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create alert action", "error", err)
				} else {
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         allocationFailedRule.ID,
						SourceActionID: nil,
						TargetActionID: alertAction.ID,
						EdgeType:       "start",
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create edge for alert action", "error", err)
					}
					log.Info(ctx, "✅ Created 'Allocation Failed - Alert Operations' rule")
				}
			}
		}

		// Rule 4: Line Item -> Check -> Reserve Pipeline
		// Demonstrates composable granular inventory actions: check_inventory -> (true_branch) -> reserve_inventory
		if orderLineItemsEntity.ID != uuid.Nil && wfEntityType.ID != uuid.Nil && onCreateTrigger.ID != uuid.Nil {
			checkReserveRule, err := busDomain.Workflow.CreateRule(ctx, workflow.NewAutomationRule{
				Name:          "Line Item Created - Check and Reserve Pipeline",
				Description:   "When a line item is created, check inventory availability then reserve if sufficient",
				EntityID:      orderLineItemsEntity.ID,
				EntityTypeID:  wfEntityType.ID,
				TriggerTypeID: onCreateTrigger.ID,
				IsActive:      false,
				IsDefault:     true,
				CreatedBy:     adminID,
			})
			if err != nil {
				log.Error(ctx, "Failed to create Check-Reserve pipeline rule", "error", err)
			} else {
				// Action 1: check_inventory (branch action)
				checkConfig := map[string]interface{}{
					"source_from_line_item": true,
					"threshold":             1,
				}
				checkConfigJSON, _ := json.Marshal(checkConfig)

				checkAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: checkReserveRule.ID,
					Name:             "Check Inventory Availability",
					Description:      "Check if inventory is available for the line item product",
					ActionConfig:     json.RawMessage(checkConfigJSON),
					IsActive:         true,
					TemplateID:       &checkInventoryTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create check_inventory action", "error", err)
				} else {
					// Action 2: reserve_inventory (on true_branch)
					reserveConfig := map[string]interface{}{
						"source_from_line_item":      true,
						"allocation_strategy":        "fifo",
						"reservation_duration_hours": 24,
						"allow_partial":              false,
					}
					reserveConfigJSON, _ := json.Marshal(reserveConfig)

					reserveAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
						AutomationRuleID: checkReserveRule.ID,
						Name:             "Reserve Inventory",
						Description:      "Reserve inventory for the line item product",
						ActionConfig:     json.RawMessage(reserveConfigJSON),
						IsActive:         true,
						TemplateID:       &reserveInventoryTemplate.ID,
					})
					if err != nil {
						log.Error(ctx, "Failed to create reserve_inventory action", "error", err)
					} else {
						// Edges: start -> check, check --(true_branch)--> reserve
						_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
							RuleID:         checkReserveRule.ID,
							SourceActionID: nil,
							TargetActionID: checkAction.ID,
							EdgeType:       "start",
							EdgeOrder:      0,
						})
						if err != nil {
							log.Error(ctx, "Failed to create start edge for check action", "error", err)
						}

						sufficientOutput := "sufficient"
						_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
							RuleID:         checkReserveRule.ID,
							SourceActionID: &checkAction.ID,
							TargetActionID: reserveAction.ID,
							EdgeType:       "sequence",
							SourceOutput:   &sufficientOutput,
							EdgeOrder:      0,
						})
						if err != nil {
							log.Error(ctx, "Failed to create sufficient output edge for reserve action", "error", err)
						}

						log.Info(ctx, "Created 'Line Item Created - Check and Reserve Pipeline' rule")
					}
				}
			}
		}

		// Rule 5: Granular Inventory Pipeline (active replacement for Rules 1-4)
		// Graph: start -> check_inventory
		//          ├── [sufficient]    -> reserve_inventory -> success_alert -> check_reorder_point
		//          │                                                             ├── [needs_reorder] -> low_stock_alert
		//          │                                                             └── [ok] -> (end)
		//          └── [insufficient]  -> insufficient_stock_alert
		if checkInventoryTemplate.ID != uuid.Nil && reserveInventoryTemplate.ID != uuid.Nil &&
			createAlertTemplate.ID != uuid.Nil && checkReorderPointTemplate.ID != uuid.Nil {

			granularRule, err := busDomain.Workflow.CreateRule(ctx, workflow.NewAutomationRule{
				Name:          "Line Item Created - Granular Inventory Pipeline",
				Description:   "When an order line item is created, check inventory, reserve if available, alert on success/failure, and check reorder point",
				EntityID:      orderLineItemsEntity.ID,
				EntityTypeID:  wfEntityType.ID,
				TriggerTypeID: onCreateTrigger.ID,
				IsActive:      true,
				IsDefault:     true,
				CreatedBy:     adminID,
			})
			if err != nil {
				log.Error(ctx, "Failed to create Granular Inventory Pipeline rule", "error", err)
			} else {
				// Action 1: check_inventory (branch action - entry point)
				checkConfig := map[string]interface{}{
					"source_from_line_item": true,
					"threshold":             1,
				}
				checkConfigJSON, _ := json.Marshal(checkConfig)

				checkAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: granularRule.ID,
					Name:             "Check Inventory Availability",
					Description:      "Check if sufficient inventory exists for the line item product",
					ActionConfig:     json.RawMessage(checkConfigJSON),
					IsActive:         true,
					TemplateID:       &checkInventoryTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create check_inventory action for granular pipeline", "error", err)
				}

				// Action 2: reserve_inventory (on true_branch from check)
				reserveConfig := map[string]interface{}{
					"source_from_line_item":      true,
					"allocation_strategy":        "fifo",
					"reservation_duration_hours": 24,
					"allow_partial":              false,
					"reference_type":             "order",
				}
				reserveConfigJSON, _ := json.Marshal(reserveConfig)

				reserveAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: granularRule.ID,
					Name:             "Reserve Inventory",
					Description:      "Reserve inventory for the line item product using FIFO strategy",
					ActionConfig:     json.RawMessage(reserveConfigJSON),
					IsActive:         true,
					TemplateID:       &reserveInventoryTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create reserve_inventory action for granular pipeline", "error", err)
				}

				// Action 3: success alert (sequence after reserve)
				successAlertCfg := map[string]interface{}{
					"alert_type": "inventory_reservation_success",
					"severity":   "critical",
					"title":      "Inventory Reservation Success",
					"message":    "Inventory has been successfully reserved for the order line item",
					"recipients": map[string]interface{}{
						"users": []string{"5cf37266-3473-4006-984f-9325122678b7"}, // Admin Gopher
						"roles": []string{},
					},
				}
				successAlertCfgJSON, _ := json.Marshal(successAlertCfg)

				successAlertAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: granularRule.ID,
					Name:             "Alert - Reservation Success",
					Description:      "Send critical alert to admin on successful inventory reservation",
					ActionConfig:     json.RawMessage(successAlertCfgJSON),
					IsActive:         true,
					TemplateID:       &createAlertTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create success alert action for granular pipeline", "error", err)
				}

				// Action 4: check_reorder_point (sequence after success alert)
				reorderCheckConfig := map[string]interface{}{
					"source_from_line_item": true,
				}
				reorderCheckConfigJSON, _ := json.Marshal(reorderCheckConfig)

				reorderCheckAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: granularRule.ID,
					Name:             "Check Reorder Point",
					Description:      "Check if inventory has fallen below the reorder point after reservation",
					ActionConfig:     json.RawMessage(reorderCheckConfigJSON),
					IsActive:         true,
					TemplateID:       &checkReorderPointTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create check_reorder_point action for granular pipeline", "error", err)
				}

				// Action 5: low stock alert (true_branch from reorder check)
				lowStockAlertCfg := map[string]interface{}{
					"alert_type": "inventory_low_stock_warning",
					"severity":   "critical",
					"title":      "Low Stock Warning",
					"message":    "Inventory has fallen below the reorder point after reservation",
					"recipients": map[string]interface{}{
						"users": []string{"5cf37266-3473-4006-984f-9325122678b7"}, // Admin Gopher
						"roles": []string{},
					},
				}
				lowStockAlertCfgJSON, _ := json.Marshal(lowStockAlertCfg)

				lowStockAlertAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: granularRule.ID,
					Name:             "Alert - Low Stock Warning",
					Description:      "Alert operations that inventory is below reorder point",
					ActionConfig:     json.RawMessage(lowStockAlertCfgJSON),
					IsActive:         true,
					TemplateID:       &createAlertTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create low stock alert action for granular pipeline", "error", err)
				}

				// Action 6: insufficient stock alert (false_branch from check)
				insufficientAlertCfg := map[string]interface{}{
					"alert_type": "inventory_insufficient_stock",
					"severity":   "critical",
					"title":      "Insufficient Stock - Reservation Failed",
					"message":    "Inventory check failed - insufficient stock available for the order line item",
					"recipients": map[string]interface{}{
						"users": []string{"5cf37266-3473-4006-984f-9325122678b7"}, // Admin Gopher
						"roles": []string{},
					},
				}
				insufficientAlertCfgJSON, _ := json.Marshal(insufficientAlertCfg)

				insufficientAlertAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: granularRule.ID,
					Name:             "Alert - Insufficient Stock",
					Description:      "Alert operations that inventory is insufficient for the order",
					ActionConfig:     json.RawMessage(insufficientAlertCfgJSON),
					IsActive:         true,
					TemplateID:       &createAlertTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create insufficient stock alert action for granular pipeline", "error", err)
				}

				// Create edges for the workflow graph
				if checkAction.ID != uuid.Nil && reserveAction.ID != uuid.Nil &&
					successAlertAction.ID != uuid.Nil && reorderCheckAction.ID != uuid.Nil &&
					lowStockAlertAction.ID != uuid.Nil && insufficientAlertAction.ID != uuid.Nil {

					// Edge 1: start -> check_inventory
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         granularRule.ID,
						SourceActionID: nil,
						TargetActionID: checkAction.ID,
						EdgeType:       "start",
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create start edge for granular pipeline", "error", err)
					}

					// Edge 2: check_inventory --[sufficient]--> reserve_inventory
					sufficientOutput := "sufficient"
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         granularRule.ID,
						SourceActionID: &checkAction.ID,
						TargetActionID: reserveAction.ID,
						EdgeType:       "sequence",
						SourceOutput:   &sufficientOutput,
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create sufficient output edge for granular pipeline", "error", err)
					}

					// Edge 3: check_inventory --[insufficient]--> insufficient_stock_alert
					insufficientOutput := "insufficient"
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         granularRule.ID,
						SourceActionID: &checkAction.ID,
						TargetActionID: insufficientAlertAction.ID,
						EdgeType:       "sequence",
						SourceOutput:   &insufficientOutput,
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create insufficient output edge for granular pipeline", "error", err)
					}

					// Edge 4: reserve_inventory --sequence--> success_alert
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         granularRule.ID,
						SourceActionID: &reserveAction.ID,
						TargetActionID: successAlertAction.ID,
						EdgeType:       "sequence",
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create sequence edge for success alert", "error", err)
					}

					// Edge 5: success_alert --sequence--> check_reorder_point
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         granularRule.ID,
						SourceActionID: &successAlertAction.ID,
						TargetActionID: reorderCheckAction.ID,
						EdgeType:       "sequence",
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create sequence edge for reorder check", "error", err)
					}

					// Edge 6: check_reorder_point --[needs_reorder]--> low_stock_alert
					needsReorderOutput := "needs_reorder"
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         granularRule.ID,
						SourceActionID: &reorderCheckAction.ID,
						TargetActionID: lowStockAlertAction.ID,
						EdgeType:       "sequence",
						SourceOutput:   &needsReorderOutput,
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create needs_reorder output edge for low stock alert", "error", err)
					}

					log.Info(ctx, "Created 'Line Item Created - Granular Inventory Pipeline' rule with 6 actions and 6 edges")
				}
			}
		}

		// =============================================================================
		// NEW DEFAULT WORKFLOWS ACROSS DOMAINS
		// =============================================================================

		// Query on_update trigger type (on_create already queried above)
		onUpdateTrigger, err := busDomain.Workflow.QueryTriggerTypeByName(ctx, "on_update")
		if err != nil {
			log.Error(ctx, "Failed to query on_update trigger type", "error", err)
		}

		// Query entities needed for new default workflows
		dwInventoryItemsEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "inventory_items")
		if err != nil {
			log.Error(ctx, "Failed to query inventory_items entity for default workflows", "error", err)
		}

		dwOrdersEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "orders")
		if err != nil {
			log.Error(ctx, "Failed to query orders entity for default workflows", "error", err)
		}

		dwUsersEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "users")
		if err != nil {
			log.Error(ctx, "Failed to query users entity for default workflows", "error", err)
		}

		dwSuppliersEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "suppliers")
		if err != nil {
			log.Error(ctx, "Failed to query suppliers entity for default workflows", "error", err)
		}

		dwSupplierProductsEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "supplier_products")
		if err != nil {
			log.Error(ctx, "Failed to query supplier_products entity for default workflows", "error", err)
		}

		dwUserAssetsEntity, err := busDomain.Workflow.QueryEntityByName(ctx, "user_assets")
		if err != nil {
			log.Error(ctx, "Failed to query user_assets entity for default workflows", "error", err)
		}

		// --- Default Workflow 1: Low Stock Alert Pipeline (inventory_items, on_update) ---
		if dwInventoryItemsEntity.ID != uuid.Nil && onUpdateTrigger.ID != uuid.Nil &&
			checkReorderPointTemplate.ID != uuid.Nil && createAlertTemplate.ID != uuid.Nil {

			lowStockRule, err := busDomain.Workflow.CreateRule(ctx, workflow.NewAutomationRule{
				Name:          "Low Stock Alert Pipeline",
				Description:   "When an inventory item is updated, check if stock is below reorder point and alert operations",
				EntityID:      dwInventoryItemsEntity.ID,
				EntityTypeID:  wfEntityType.ID,
				TriggerTypeID: onUpdateTrigger.ID,
				IsActive:      false,
				IsDefault:     true,
				CreatedBy:     adminID,
			})
			if err != nil {
				log.Error(ctx, "Failed to create Low Stock Alert Pipeline rule", "error", err)
			} else {
				reorderCheckCfg, _ := json.Marshal(map[string]interface{}{
					"source_from_entity": true,
				})
				checkAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: lowStockRule.ID,
					Name:             "Check Reorder Point",
					Description:      "Check if inventory is below reorder point",
					ActionConfig:     json.RawMessage(reorderCheckCfg),
					IsActive:         true,
					TemplateID:       &checkReorderPointTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create check_reorder_point action for low stock pipeline", "error", err)
				} else {
					alertCfg, _ := json.Marshal(map[string]interface{}{
						"alert_type": "low_stock_warning",
						"severity":   "warning",
						"title":      "Low Stock Alert",
						"message":    "Inventory item has fallen below the reorder point",
						"recipients": map[string]interface{}{
							"users": []string{"5cf37266-3473-4006-984f-9325122678b7"},
							"roles": []string{},
						},
					})
					alertAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
						AutomationRuleID: lowStockRule.ID,
						Name:             "Create Low Stock Alert",
						Description:      "Alert operations about low stock levels",
						ActionConfig:     json.RawMessage(alertCfg),
						IsActive:         true,
						TemplateID:       &createAlertTemplate.ID,
					})
					if err != nil {
						log.Error(ctx, "Failed to create alert action for low stock pipeline", "error", err)
					} else {
						// start -> check_reorder_point
						_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
							RuleID:         lowStockRule.ID,
							SourceActionID: nil,
							TargetActionID: checkAction.ID,
							EdgeType:       "start",
							EdgeOrder:      0,
						})
						if err != nil {
							log.Error(ctx, "Failed to create start edge for low stock pipeline", "error", err)
						}
						// check_reorder_point --[needs_reorder]--> create_alert
						needsReorder := "needs_reorder"
						_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
							RuleID:         lowStockRule.ID,
							SourceActionID: &checkAction.ID,
							TargetActionID: alertAction.ID,
							EdgeType:       "sequence",
							SourceOutput:   &needsReorder,
							EdgeOrder:      0,
						})
						if err != nil {
							log.Error(ctx, "Failed to create sequence edge for low stock pipeline", "error", err)
						}
						log.Info(ctx, "Created 'Low Stock Alert Pipeline' default workflow")
					}
				}
			}
		}

		// --- Default Workflow 2: Item Created - Log Audit Entry (inventory_items, on_create) ---
		if dwInventoryItemsEntity.ID != uuid.Nil && logAuditEntryTemplate.ID != uuid.Nil {
			auditRule, err := busDomain.Workflow.CreateRule(ctx, workflow.NewAutomationRule{
				Name:          "Item Created - Log Audit Entry",
				Description:   "When an inventory item is created, log an audit trail entry",
				EntityID:      dwInventoryItemsEntity.ID,
				EntityTypeID:  wfEntityType.ID,
				TriggerTypeID: onCreateTrigger.ID,
				IsActive:      false,
				IsDefault:     true,
				CreatedBy:     adminID,
			})
			if err != nil {
				log.Error(ctx, "Failed to create Item Created - Log Audit Entry rule", "error", err)
			} else {
				auditCfg, _ := json.Marshal(map[string]interface{}{
					"log_level": "info",
					"message":   "New inventory item created",
				})
				auditAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: auditRule.ID,
					Name:             "Log Audit Entry",
					Description:      "Write audit trail for new inventory item",
					ActionConfig:     json.RawMessage(auditCfg),
					IsActive:         true,
					TemplateID:       &logAuditEntryTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create audit action for item created rule", "error", err)
				} else {
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         auditRule.ID,
						SourceActionID: nil,
						TargetActionID: auditAction.ID,
						EdgeType:       "start",
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create start edge for item created audit rule", "error", err)
					}
					log.Info(ctx, "Created 'Item Created - Log Audit Entry' default workflow")
				}
			}
		}

		// --- Default Workflow 3: Order Created - Confirmation Alert (orders, on_create) ---
		if dwOrdersEntity.ID != uuid.Nil && createAlertTemplate.ID != uuid.Nil {
			orderAlertRule, err := busDomain.Workflow.CreateRule(ctx, workflow.NewAutomationRule{
				Name:          "Order Created - Confirmation Alert",
				Description:   "When a sales order is created, send a confirmation alert",
				EntityID:      dwOrdersEntity.ID,
				EntityTypeID:  wfEntityType.ID,
				TriggerTypeID: onCreateTrigger.ID,
				IsActive:      false,
				IsDefault:     true,
				CreatedBy:     adminID,
			})
			if err != nil {
				log.Error(ctx, "Failed to create Order Created - Confirmation Alert rule", "error", err)
			} else {
				alertCfg, _ := json.Marshal(map[string]interface{}{
					"alert_type": "order_confirmation",
					"severity":   "info",
					"title":      "New Order Created",
					"message":    "A new sales order has been created",
					"recipients": map[string]interface{}{
						"users": []string{"5cf37266-3473-4006-984f-9325122678b7"},
						"roles": []string{},
					},
				})
				alertAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: orderAlertRule.ID,
					Name:             "Create Order Confirmation Alert",
					Description:      "Send confirmation alert for new order",
					ActionConfig:     json.RawMessage(alertCfg),
					IsActive:         true,
					TemplateID:       &createAlertTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create alert action for order confirmation rule", "error", err)
				} else {
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         orderAlertRule.ID,
						SourceActionID: nil,
						TargetActionID: alertAction.ID,
						EdgeType:       "start",
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create start edge for order confirmation rule", "error", err)
					}
					log.Info(ctx, "Created 'Order Created - Confirmation Alert' default workflow")
				}
			}
		}

		// --- Default Workflow 4: Order Updated - Notify Operations (orders, on_update) ---
		if dwOrdersEntity.ID != uuid.Nil && onUpdateTrigger.ID != uuid.Nil && createAlertTemplate.ID != uuid.Nil {
			orderUpdateRule, err := busDomain.Workflow.CreateRule(ctx, workflow.NewAutomationRule{
				Name:          "Order Updated - Notify Operations",
				Description:   "When a sales order is updated, notify the operations team",
				EntityID:      dwOrdersEntity.ID,
				EntityTypeID:  wfEntityType.ID,
				TriggerTypeID: onUpdateTrigger.ID,
				IsActive:      false,
				IsDefault:     true,
				CreatedBy:     adminID,
			})
			if err != nil {
				log.Error(ctx, "Failed to create Order Updated - Notify Operations rule", "error", err)
			} else {
				alertCfg, _ := json.Marshal(map[string]interface{}{
					"alert_type": "order_updated",
					"severity":   "info",
					"title":      "Order Updated",
					"message":    "A sales order has been updated",
					"recipients": map[string]interface{}{
						"users": []string{"5cf37266-3473-4006-984f-9325122678b7"},
						"roles": []string{},
					},
				})
				alertAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: orderUpdateRule.ID,
					Name:             "Create Order Update Alert",
					Description:      "Notify operations about order update",
					ActionConfig:     json.RawMessage(alertCfg),
					IsActive:         true,
					TemplateID:       &createAlertTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create alert action for order update rule", "error", err)
				} else {
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         orderUpdateRule.ID,
						SourceActionID: nil,
						TargetActionID: alertAction.ID,
						EdgeType:       "start",
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create start edge for order update rule", "error", err)
					}
					log.Info(ctx, "Created 'Order Updated - Notify Operations' default workflow")
				}
			}
		}

		// --- Default Workflow 5: New User Created - Welcome Alert (users, on_create) ---
		if dwUsersEntity.ID != uuid.Nil && createAlertTemplate.ID != uuid.Nil {
			welcomeRule, err := busDomain.Workflow.CreateRule(ctx, workflow.NewAutomationRule{
				Name:          "New User Created - Welcome Alert",
				Description:   "When a new user is created, send a welcome alert to the team",
				EntityID:      dwUsersEntity.ID,
				EntityTypeID:  wfEntityType.ID,
				TriggerTypeID: onCreateTrigger.ID,
				IsActive:      false,
				IsDefault:     true,
				CreatedBy:     adminID,
			})
			if err != nil {
				log.Error(ctx, "Failed to create New User Created - Welcome Alert rule", "error", err)
			} else {
				alertCfg, _ := json.Marshal(map[string]interface{}{
					"alert_type": "new_user_welcome",
					"severity":   "info",
					"title":      "New User Created",
					"message":    "A new user has been added to the system",
					"recipients": map[string]interface{}{
						"users": []string{"5cf37266-3473-4006-984f-9325122678b7"},
						"roles": []string{},
					},
				})
				alertAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: welcomeRule.ID,
					Name:             "Create Welcome Alert",
					Description:      "Send welcome alert for new user",
					ActionConfig:     json.RawMessage(alertCfg),
					IsActive:         true,
					TemplateID:       &createAlertTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create alert action for welcome rule", "error", err)
				} else {
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         welcomeRule.ID,
						SourceActionID: nil,
						TargetActionID: alertAction.ID,
						EdgeType:       "start",
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create start edge for welcome rule", "error", err)
					}
					log.Info(ctx, "Created 'New User Created - Welcome Alert' default workflow")
				}
			}
		}

		// --- Default Workflow 6: Supplier Added - Alert Team (suppliers, on_create) ---
		if dwSuppliersEntity.ID != uuid.Nil && createAlertTemplate.ID != uuid.Nil {
			supplierRule, err := busDomain.Workflow.CreateRule(ctx, workflow.NewAutomationRule{
				Name:          "Supplier Added - Alert Team",
				Description:   "When a new supplier is added, alert the procurement team",
				EntityID:      dwSuppliersEntity.ID,
				EntityTypeID:  wfEntityType.ID,
				TriggerTypeID: onCreateTrigger.ID,
				IsActive:      false,
				IsDefault:     true,
				CreatedBy:     adminID,
			})
			if err != nil {
				log.Error(ctx, "Failed to create Supplier Added - Alert Team rule", "error", err)
			} else {
				alertCfg, _ := json.Marshal(map[string]interface{}{
					"alert_type": "new_supplier",
					"severity":   "info",
					"title":      "New Supplier Added",
					"message":    "A new supplier has been added to the system",
					"recipients": map[string]interface{}{
						"users": []string{"5cf37266-3473-4006-984f-9325122678b7"},
						"roles": []string{},
					},
				})
				alertAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: supplierRule.ID,
					Name:             "Create Supplier Alert",
					Description:      "Alert procurement team about new supplier",
					ActionConfig:     json.RawMessage(alertCfg),
					IsActive:         true,
					TemplateID:       &createAlertTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create alert action for supplier rule", "error", err)
				} else {
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         supplierRule.ID,
						SourceActionID: nil,
						TargetActionID: alertAction.ID,
						EdgeType:       "start",
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create start edge for supplier rule", "error", err)
					}
					log.Info(ctx, "Created 'Supplier Added - Alert Team' default workflow")
				}
			}
		}

		// --- Default Workflow 7: Supplier Product Added - Log Audit (supplier_products, on_create) ---
		if dwSupplierProductsEntity.ID != uuid.Nil && logAuditEntryTemplate.ID != uuid.Nil {
			spAuditRule, err := busDomain.Workflow.CreateRule(ctx, workflow.NewAutomationRule{
				Name:          "Supplier Product Added - Log Audit",
				Description:   "When a supplier product is added, log an audit trail entry",
				EntityID:      dwSupplierProductsEntity.ID,
				EntityTypeID:  wfEntityType.ID,
				TriggerTypeID: onCreateTrigger.ID,
				IsActive:      false,
				IsDefault:     true,
				CreatedBy:     adminID,
			})
			if err != nil {
				log.Error(ctx, "Failed to create Supplier Product Added - Log Audit rule", "error", err)
			} else {
				auditCfg, _ := json.Marshal(map[string]interface{}{
					"log_level": "info",
					"message":   "New supplier product added",
				})
				auditAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: spAuditRule.ID,
					Name:             "Log Audit Entry",
					Description:      "Write audit trail for new supplier product",
					ActionConfig:     json.RawMessage(auditCfg),
					IsActive:         true,
					TemplateID:       &logAuditEntryTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create audit action for supplier product rule", "error", err)
				} else {
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         spAuditRule.ID,
						SourceActionID: nil,
						TargetActionID: auditAction.ID,
						EdgeType:       "start",
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create start edge for supplier product audit rule", "error", err)
					}
					log.Info(ctx, "Created 'Supplier Product Added - Log Audit' default workflow")
				}
			}
		}

		// --- Default Workflow 8: Asset Assigned - Send Notification (user_assets, on_create) ---
		if dwUserAssetsEntity.ID != uuid.Nil && sendNotificationTemplate.ID != uuid.Nil {
			assetRule, err := busDomain.Workflow.CreateRule(ctx, workflow.NewAutomationRule{
				Name:          "Asset Assigned - Send Notification",
				Description:   "When an asset is assigned to a user, send a notification",
				EntityID:      dwUserAssetsEntity.ID,
				EntityTypeID:  wfEntityType.ID,
				TriggerTypeID: onCreateTrigger.ID,
				IsActive:      false,
				IsDefault:     true,
				CreatedBy:     adminID,
			})
			if err != nil {
				log.Error(ctx, "Failed to create Asset Assigned - Send Notification rule", "error", err)
			} else {
				notifCfg, _ := json.Marshal(map[string]interface{}{
					"channel":  "in_app",
					"title":    "Asset Assigned",
					"message":  "An asset has been assigned to you",
					"priority": "normal",
				})
				notifAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
					AutomationRuleID: assetRule.ID,
					Name:             "Send Asset Assignment Notification",
					Description:      "Notify user about asset assignment",
					ActionConfig:     json.RawMessage(notifCfg),
					IsActive:         true,
					TemplateID:       &sendNotificationTemplate.ID,
				})
				if err != nil {
					log.Error(ctx, "Failed to create notification action for asset assignment rule", "error", err)
				} else {
					_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
						RuleID:         assetRule.ID,
						SourceActionID: nil,
						TargetActionID: notifAction.ID,
						EdgeType:       "start",
						EdgeOrder:      0,
					})
					if err != nil {
						log.Error(ctx, "Failed to create start edge for asset assignment rule", "error", err)
					}
					log.Info(ctx, "Created 'Asset Assigned - Send Notification' default workflow")
				}
			}
		}
	}

	log.Info(ctx, "Workflow automation rules seeding complete")

	return nil
}
