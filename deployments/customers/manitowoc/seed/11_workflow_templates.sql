BEGIN;
--
-- PostgreSQL database dump
--

\restrict 5ah4wSfHAiJbl35miy66J2mCGJrSyGuOiUcMbWNzwnc3qDgbTyA1azdDP1UlqAN

-- Dumped from database version 16.4 (Debian 16.4-1.pgdg120+2)
-- Dumped by pg_dump version 17.6

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: action_templates; Type: TABLE DATA; Schema: workflow; Owner: -
--

INSERT INTO workflow.action_templates VALUES ('243719aa-fea6-4ebc-951e-77f43014a74b', 'Allocate Inventory', 'Allocates inventory for an order or request', 'allocate_inventory', 'material-symbols:inventory-2', '{}', '2026-03-19 15:16:56.253901', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('c4496dcf-91c2-4518-b8ef-58c12ede031e', 'Update Field', 'Updates a field on the target entity', 'update_field', 'material-symbols:edit-note', '{}', '2026-03-19 15:16:56.256299', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('38262dc8-bf9d-4103-a2ad-6532ae0c0fbb', 'Create Alert', 'Creates an alert notification', 'create_alert', 'material-symbols:notification-important', '{}', '2026-03-19 15:16:56.258325', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('89a968aa-f318-4609-a8d9-ca806c45ae72', 'Check Inventory', 'Checks inventory availability against a threshold', 'check_inventory', 'material-symbols:fact-check', '{"threshold": 1}', '2026-03-19 15:16:56.260224', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('d5bfa1de-b6ff-4072-83a4-edea6df9d212', 'Reserve Inventory', 'Reserves inventory with idempotency support', 'reserve_inventory', 'material-symbols:bookmark-added', '{"allow_partial": false, "allocation_strategy": "fifo", "reservation_duration_hours": 24}', '2026-03-19 15:16:56.262341', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('85c57856-2a58-4af1-a420-ed3aecdb7417', 'Release Reservation', 'Releases reserved inventory back to available stock', 'release_reservation', 'material-symbols:remove-shopping-cart', '{}', '2026-03-19 15:16:56.264241', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('2f40af86-53e1-4247-bee0-2bb27ce62515', 'Commit Allocation', 'Commits reserved inventory to allocated status', 'commit_allocation', 'material-symbols:check-circle', '{}', '2026-03-19 15:16:56.266258', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('ff1dfb3e-2457-453f-b30e-6c72a1a6aaca', 'Check Reorder Point', 'Checks if inventory is below reorder point', 'check_reorder_point', 'material-symbols:trending-down', '{}', '2026-03-19 15:16:56.268207', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('c709ad5d-6581-4932-9899-1db16f3d5bb7', 'Log Audit Entry', 'Write an audit trail entry to the workflow audit log', 'log_audit_entry', 'material-symbols:history-edu', '{}', '2026-03-19 15:16:56.270126', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('54f12830-7e38-476d-824d-ed0bb087b17f', 'Create Entity', 'Create a new entity record in the database', 'create_entity', 'material-symbols:note-add', '{}', '2026-03-19 15:16:56.272005', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('fc7357c1-05b8-43f5-a330-ddcdc95aa425', 'Delay', 'Pause workflow execution for a specified duration', 'delay', 'material-symbols:timer', '{"duration": "5m"}', '2026-03-19 15:16:56.273911', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('b390967b-d3ae-4c2f-bf98-d04bfee41ecd', 'Evaluate Condition', 'Evaluates conditions and determines branch direction', 'evaluate_condition', 'material-symbols:fork-right', '{}', '2026-03-19 15:16:56.27578', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('a42018f0-b03b-4d2c-8888-f8ff326e8f77', 'Lookup Entity', 'Look up entity data by filter criteria', 'lookup_entity', 'material-symbols:manage-search', '{}', '2026-03-19 15:16:56.277665', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('3098ccb2-f02b-4b35-8251-fe0b6a6fec6c', 'Seek Approval', 'Creates an approval request for specified users', 'seek_approval', 'material-symbols:approval', '{}', '2026-03-19 15:16:56.279513', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('adb05ce1-74fb-4299-8025-8f6a0e810091', 'Send Email', 'Sends an email to specified recipients', 'send_email', 'material-symbols:forward-to-inbox', '{}', '2026-03-19 15:16:56.28139', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('ace2438c-1ffa-41ca-958a-c204b016672e', 'Send Notification', 'Sends in-app notifications through various channels', 'send_notification', 'material-symbols:campaign', '{}', '2026-03-19 15:16:56.283278', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('67a4db1f-165d-4072-87a2-e172ecc37998', 'Transition Status', 'Transition an entity field from one status to another', 'transition_status', 'material-symbols:swap-horiz', '{}', '2026-03-19 15:16:56.285192', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('36876141-0ff7-4219-8afa-6cbc80ac9e1e', 'Call Webhook', 'Makes an outbound HTTP request to an external URL', 'call_webhook', 'material-symbols:webhook', '{"method": "POST", "timeout_seconds": 30}', '2026-03-19 15:16:56.287179', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('97360a43-23c9-4093-85de-acb8bf8d7958', 'Receive Inventory', 'Receives inventory into a warehouse location from a purchase order', 'receive_inventory', 'material-symbols:local-shipping', '{"source_from_po": true}', '2026-03-19 15:16:56.289178', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('f454f603-2e66-4374-91c8-9f57872175d5', 'Create Purchase Order', 'Creates a purchase order with line items for supplier procurement', 'create_purchase_order', 'material-symbols:receipt-long', '{}', '2026-03-19 15:16:56.291252', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);
INSERT INTO workflow.action_templates VALUES ('64f08872-ca2f-444f-adc8-b45a6ec6f46d', 'Create Put-Away Task', 'Creates a put-away task directing floor workers to shelve received goods', 'create_put_away_task', 'material-symbols:shelves', '{"source_from_po": true, "reference_number": "PO-RCV-{{purchase_order_id}}", "location_strategy": "po_delivery"}', '2026-03-19 15:16:56.29359', 'b4f7bc0c-9d8e-4834-a236-1204ae3d7254', true, NULL);


--
-- PostgreSQL database dump complete
--

\unrestrict 5ah4wSfHAiJbl35miy66J2mCGJrSyGuOiUcMbWNzwnc3qDgbTyA1azdDP1UlqAN

COMMIT;
