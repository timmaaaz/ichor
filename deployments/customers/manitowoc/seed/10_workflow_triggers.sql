BEGIN;
--
-- PostgreSQL database dump
--

\restrict mm4lWKgDNxhmSgxu7CHfazNM5y73oc9mdbDNrNkSxaRxcHrZr6actqvjiqKIgRU

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
-- Data for Name: entity_types; Type: TABLE DATA; Schema: workflow; Owner: -
--

INSERT INTO workflow.entity_types VALUES ('0b290c2f-8259-4d11-91f7-ff333b7350ba', 'table', 'Database table entity', true, NULL, NULL);
INSERT INTO workflow.entity_types VALUES ('9287fa5f-fb45-4c8b-8339-e5fe527997a9', 'view', 'Database view entity', true, NULL, NULL);


--
-- Data for Name: entities; Type: TABLE DATA; Schema: workflow; Owner: -
--

INSERT INTO workflow.entities VALUES ('cb7a5707-c883-4edd-8f6e-8e326e30e01e', 'enum_labels', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'config', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('f68302bb-0eeb-4db0-82af-1e18e70fa87c', 'serial_numbers', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'inventory', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('b78e7fd1-4096-43b0-9324-d06fa0fcc2bb', 'approval_status', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'assets', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('b99b0495-f19d-49e4-b4e1-f665e371962d', 'audit_log_2026_03', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('fb6e19e0-f140-469d-8b3a-90bbf07d55a2', 'inventory_locations', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'inventory', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('6be95de6-8a85-44df-b5cb-33618c80bfa4', 'zones', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'inventory', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('e6c404be-315a-4255-a8bc-93b36bc004ab', 'user_approval_comments', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'hr', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('752c09a1-0f66-4e80-999a-8e5e4f0f59d9', 'user_assets', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'assets', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('869041b3-abe6-49ce-889d-1b522e71daa2', 'order_fulfillment_statuses', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'sales', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('ddf30afe-4c25-4d0f-b520-fb265dd3e9bc', 'warehouses', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'inventory', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('bb1583ed-a5c9-4428-b0ba-2bee612eaf45', 'purchase_order_line_items', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'procurement', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('cff4b553-9418-4528-9785-7e84ee698901', 'asset_types', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'assets', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('0cef90ec-1043-4012-ba65-41d78ca70920', 'contact_infos', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'core', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('80cb411c-cf3c-4d11-b005-d3bc8651f4cc', 'table_configs', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'config', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('02c16263-e046-4b38-af1f-34aa8c3bc0a7', 'homes', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'hr', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('53c26f70-0f8c-487c-942c-ee7a9c2721bf', 'pages', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'core', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('56269ad6-7768-44a7-a647-189b892cf6c2', 'users', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'core', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('29a6fcaa-1073-4bb1-96a2-5f8af0abfb62', 'audit_log_2026_02', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('fc939d73-ac1f-4592-a78d-d3d2992a649c', 'streets', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'geography', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('ad8a131e-bb64-40b7-adad-00caa425ba08', 'action_templates', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('b613efff-de65-4810-8296-fe2f190e387c', 'timezones', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'geography', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('5d869b21-2955-4864-948e-8af84460eb17', 'customers', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'sales', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('f2f92034-f626-405a-9279-1af798771d76', 'asset_tags', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'assets', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('4aea30f6-778d-4671-b6d2-330a828e4bf1', 'brands', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'products', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('13d5097f-8b38-4d7b-b935-744c51f5a3e8', 'automation_rules', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('dae820c3-72ba-4c29-8627-887b081fb39c', 'order_line_items', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'sales', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('5452131c-538b-4022-bcca-ed7cbb6ac82e', 'line_item_fulfillment_statuses', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'sales', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('0414fb53-8120-401f-b5f0-9818c279d5f4', 'alert_acknowledgments', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('9f80f1f8-c7a0-49b3-93eb-87e04f3e215f', 'purchase_order_statuses', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'procurement', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('3fbfeb0a-2a97-4271-87dc-4cdfbb2f6b12', 'audit_log_2026_05', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('de2d2e72-e405-4b9f-b3eb-b0d14539d48e', 'form_fields', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'config', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('7addaacb-191d-473a-9a0e-b64a045b760f', 'page_configs', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'config', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('4b4c942e-3f07-42cf-ae9d-e63ae5f14fd3', 'inventory_adjustments', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'inventory', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('2bcfb934-d0d9-4103-ab0b-9385de989cbf', 'product_uoms', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'products', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('e3e0f0a1-1b10-4035-8c38-a777c5c766fa', 'page_content', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'config', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('d6ff1acf-cb79-4f6b-bca1-262027758328', 'lot_locations', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'inventory', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('9d4133ab-ae18-4694-b684-e8f894a06c88', 'product_costs', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'products', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('3a18b77b-72f8-4b78-8742-3ce122808c6b', 'rule_actions', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('d70fe1a3-3d07-4e31-815e-47c56d178f7b', 'physical_attributes', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'products', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('32e568d9-b515-4fa9-b8c5-3ccb8c74b7fb', 'fulfillment_status', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'assets', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('2413b12b-a718-478b-887e-1ba0faebfaae', 'tags', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'assets', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('b4fc3faf-fd47-4dbd-8cf6-0ec0e83b2819', 'titles', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'hr', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('210cd1a5-6403-4c9d-94f8-307d346b8063', 'regions', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'geography', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('db950ef2-c743-4caf-9969-0308f66dc455', 'orders', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'sales', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('df0d8cbf-f1a0-4e13-ba6f-8e3e71d70c42', 'settings', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'config', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('026f9c8c-3040-4e47-94d3-4519118ac089', 'allocation_results', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('cf4907a9-30cb-4c7e-9c83-a5285bc5df4a', 'inventory_items', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'inventory', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('cfd6cb48-5711-4a79-a076-b1870da8e54d', 'inventory_transactions', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'inventory', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('1b34f4f8-f063-4807-9dce-f44f6ec5fa04', 'page_action_buttons', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'config', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('8d88fce7-936d-4da1-85fe-b6eba45249b5', 'supplier_products', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'procurement', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('7ce8f255-1f62-4ca0-8b5f-dcce16f3b180', 'countries', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'geography', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('c14081f0-e41b-43c1-bf30-07f6148fe492', 'put_away_tasks', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'inventory', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('7c891044-8eae-4cbd-8ff4-1f76be320528', 'alerts', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('fd5ae2a8-ba3a-495e-8624-3a9219342095', 'rule_dependencies', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('b4484690-7530-44e2-b8a9-06311107c954', 'automation_executions', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('aaa819f1-0d3e-4317-ad0b-d74cb96ed016', 'offices', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'hr', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('68a78e5a-ec87-4cbd-9f11-822b38a767dc', 'audit_log_default', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('13b1ab26-e90a-4701-90ca-44e9f1aceec9', 'reports_to', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'hr', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('2e7e9703-9fc6-4c83-ad2e-83d23599f95c', 'lot_trackings', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'inventory', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('4165dfe3-abd3-40c2-8754-32b92dd18345', 'action_edges', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('1d34ce7e-c28c-4868-805e-3bdb573d2122', 'trigger_types', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('01bd8c57-033e-43b2-af3a-ab22b9b5c699', 'user_approval_status', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'hr', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('305220bf-a322-4cca-8393-d3f64ba3a2f6', 'product_categories', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'products', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('aee62ed5-9907-462d-82dd-8aa85f4b1280', 'cost_history', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'products', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('bb102f81-2e16-4782-b267-ebd7b5b247cf', 'suppliers', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'procurement', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('d51bb2a7-d30d-4736-a109-e78f588c5ef8', 'cities', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'geography', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('0afb31c2-e0e9-4223-a186-b56d887fdf25', 'approval_requests', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('6ee76df8-cfcf-4db9-aaeb-0ceaed5a773d', 'entity_types', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('a9bd21a7-29b5-4498-8c72-019f6a634043', 'audit_log_2026_04', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('57a39223-3b9e-46aa-9a7d-d995007165cf', 'entities', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('8b950e9c-0e5c-4baf-82b3-88cba1d2e99d', 'notification_deliveries', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('352f211f-a7be-4022-908c-ac69c441a749', 'payment_terms', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'core', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('9510912a-72f4-496b-b3be-535010e51dc7', 'role_pages', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'core', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('cbd8ef5e-ff7c-4364-9104-eaf3e1196b0f', 'purchase_order_line_item_statuses', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'procurement', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('11fe5164-22bf-4401-ac41-e9ddb8b5dd09', 'page_actions', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'config', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('2c334142-2be4-48a7-96c1-f79b32704f5f', 'table_access', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'core', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('f564f8e4-f5a4-4a91-ad1f-98234da937ec', 'action_permissions', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('b3fbd60b-bb42-4b12-9aeb-ae5368a25f61', 'assets', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'assets', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('c57b1144-b2e4-42b2-ab8c-ee8a70404e94', 'valid_assets', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'assets', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('a49551f8-747b-4dfa-b6f9-5f4857f1825c', 'currencies', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'core', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('f1c0143d-146a-42eb-82b0-6f991413d4c1', 'purchase_orders', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'procurement', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('0d4a49b8-5623-446e-ac52-03f41d8ac6fa', 'page_action_dropdowns', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'config', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('ecf5b3f4-8b6b-45d0-9653-e6f0c82c30dd', 'quality_inspections', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'inventory', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('29506447-0774-4ec6-a2a6-c99e89cbc007', 'page_action_dropdown_items', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'config', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('c6e7c0f9-54e8-4d57-881a-61b0828631e0', 'products', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'products', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('60952993-9b4f-49bf-9345-e9856ea5c621', 'alert_recipients', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('4d5a28e5-3e03-4cb7-870d-e6c8db1def40', 'quality_metrics', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'products', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('fe8bc537-c48d-4868-a351-eaf8ef3fa233', 'asset_conditions', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'assets', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('cdf0f40f-7660-4897-943e-5397ac1ee6cf', 'roles', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'core', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('12dd3221-d4ae-4b09-ab85-ae240ac47747', 'forms', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'config', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('cfe8d570-de20-4da8-be32-4576b7f6b7ce', 'user_roles', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'core', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('90a31a29-e1bc-4f6c-8493-5ea1bb70b13b', 'transfer_orders', '0b290c2f-8259-4d11-91f7-ff333b7350ba', 'inventory', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('5b2dc366-c340-47f2-8985-27609c6d811a', 'rule_actions_view', '9287fa5f-fb45-4c8b-8339-e5fe527997a9', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('ea91fd46-35ae-447f-9ea3-a60eafce9ca0', 'automation_rules_view', '9287fa5f-fb45-4c8b-8339-e5fe527997a9', 'workflow', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('6c60e404-f4e8-4a8f-a2a8-f8ea74a23eea', 'order_line_items_base', '9287fa5f-fb45-4c8b-8339-e5fe527997a9', 'sales', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('2d09fcba-aa0b-4870-bf5c-919ac9aa253f', 'active_table_configs', '9287fa5f-fb45-4c8b-8339-e5fe527997a9', 'config', true, '2026-03-19 15:16:10.073592', NULL);
INSERT INTO workflow.entities VALUES ('2fdfc6d7-21c7-42b9-b5a7-967ef2419647', 'orders_base', '9287fa5f-fb45-4c8b-8339-e5fe527997a9', 'sales', true, '2026-03-19 15:16:10.073592', NULL);


--
-- Data for Name: trigger_types; Type: TABLE DATA; Schema: workflow; Owner: -
--

INSERT INTO workflow.trigger_types VALUES ('76753388-e752-4ba7-ae4d-9111050e7afb', 'on_create', 'Triggered when a new entity is created', true, NULL);
INSERT INTO workflow.trigger_types VALUES ('334292a9-e7a5-4066-9b85-c99891a689ee', 'on_update', 'Triggered when an existing entity is updated', true, NULL);
INSERT INTO workflow.trigger_types VALUES ('9f631f2c-cc94-47a6-a448-07272c5f73a3', 'on_delete', 'Triggered when an entity is deleted', true, NULL);
INSERT INTO workflow.trigger_types VALUES ('358d75d9-a2a9-49f0-a326-62f0a64b0684', 'scheduled', 'Triggered based on a schedule', true, NULL);


--
-- PostgreSQL database dump complete
--

\unrestrict mm4lWKgDNxhmSgxu7CHfazNM5y73oc9mdbDNrNkSxaRxcHrZr6actqvjiqKIgRU

COMMIT;
