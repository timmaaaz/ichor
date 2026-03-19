BEGIN;
--
-- PostgreSQL database dump
--


-- Dumped from database version 16.4 (Debian 16.4-1.pgdg120+2)
-- Dumped by pg_dump version 17.6

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: page_actions; Type: TABLE DATA; Schema: config; Owner: -
--

INSERT INTO config.page_actions VALUES ('9caaef40-b1b0-4391-a8c3-3ad9c38dcea0', 'c0590f21-21f2-41b8-8a45-1293f0c194a4', 'button', 1, true);
INSERT INTO config.page_actions VALUES ('c5f206d3-6f93-4435-bb37-303ce4fe5103', '2d918bfe-e839-4b6a-9c12-890f0ec37d0d', 'button', 1, true);
INSERT INTO config.page_actions VALUES ('6caca2f8-97df-4954-93d9-60cd63bedc7f', 'd1556062-a25d-428c-8841-e0c3509ab189', 'button', 1, true);
INSERT INTO config.page_actions VALUES ('441a85c4-844c-4f69-9be2-6462ce5230d2', 'd1556062-a25d-428c-8841-e0c3509ab189', 'button', 2, true);
INSERT INTO config.page_actions VALUES ('1d5ba7f0-6bf9-41e8-b126-9dd4509e3ee6', 'd1556062-a25d-428c-8841-e0c3509ab189', 'button', 3, true);
INSERT INTO config.page_actions VALUES ('e9c6327a-059e-4e70-a7b6-af376d0384b2', 'd1556062-a25d-428c-8841-e0c3509ab189', 'button', 4, true);
INSERT INTO config.page_actions VALUES ('14464878-cdad-44ff-a781-a8446706ec52', 'd1556062-a25d-428c-8841-e0c3509ab189', 'button', 5, true);
INSERT INTO config.page_actions VALUES ('64914175-c59e-4337-a530-efb304483319', 'b716ebfd-3c22-4294-81f3-4ee3e68f7b58', 'button', 1, true);
INSERT INTO config.page_actions VALUES ('f18471cd-bd55-4d03-a914-5ba22e2b241b', '339bc185-9045-4807-b463-de065529a152', 'button', 1, true);
INSERT INTO config.page_actions VALUES ('146be8e0-d43c-4e69-af6c-8792a37ac066', '570f2a2b-7898-4231-8931-5174f02f45b0', 'button', 1, true);
INSERT INTO config.page_actions VALUES ('b37fe8cd-3135-40fa-93ec-3f98c677d416', '2f7b7513-8a92-4922-acfa-41cc94e8a134', 'button', 1, true);
INSERT INTO config.page_actions VALUES ('8dd86e75-5afc-4929-bfaa-1b1613cf4db5', '47739fd5-ee02-4180-af5a-8db38bb41d5e', 'button', 1, true);
INSERT INTO config.page_actions VALUES ('9a9b7b52-c5b9-456e-825e-69be8b4d5557', '1b20927b-8682-4525-b18f-f0e7d1e1d1de', 'button', 1, true);
INSERT INTO config.page_actions VALUES ('70011703-9a8b-4803-b2df-f58f19c9afdd', 'e5a036fb-5e93-437c-a41e-b184db8cbaee', 'button', 1, true);
INSERT INTO config.page_actions VALUES ('a1107ca5-3b38-4f36-95a0-f5786966ae18', 'aca655fd-1b0d-40b0-b02c-be56ebb7916e', 'button', 1, true);
INSERT INTO config.page_actions VALUES ('0eb71ace-c6e2-4788-b91c-8e168bbe5744', 'd25ec547-eb68-4847-8864-cf0d6532b913', 'button', 1, true);
INSERT INTO config.page_actions VALUES ('e2f67272-0037-4eb3-beba-1b8ee7c35298', 'd25ec547-eb68-4847-8864-cf0d6532b913', 'button', 2, true);
INSERT INTO config.page_actions VALUES ('02ee40f0-c1d5-4fcd-a65f-6fb4550d5299', '592e7016-6e68-413b-a354-17ec6c38da83', 'button', 1, true);
INSERT INTO config.page_actions VALUES ('1dbe4633-46ca-4b35-b4a2-cbb598be1791', '193aabb2-e3b0-474d-b86f-8409415d0b11', 'button', 1, true);
INSERT INTO config.page_actions VALUES ('c35ed9f7-6b99-4377-8360-e0f903be7060', '631d9418-04de-4c4f-8eae-3269f114c60b', 'button', 1, true);


--
-- Data for Name: page_action_buttons; Type: TABLE DATA; Schema: config; Owner: -
--

INSERT INTO config.page_action_buttons VALUES ('9caaef40-b1b0-4391-a8c3-3ad9c38dcea0', 'New Item', 'material-symbols:add-box', '/inventory/items/new', 'default', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('c5f206d3-6f93-4435-bb37-303ce4fe5103', 'New Supplier', 'material-symbols:add-business', '/procurement/suppliers/new', 'default', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('6caca2f8-97df-4954-93d9-60cd63bedc7f', 'Default Button', 'material-symbols:add-circle', '/test/default', 'default', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('441a85c4-844c-4f69-9be2-6462ce5230d2', 'Secondary Button', 'material-symbols:edit', '/test/secondary', 'secondary', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('1d5ba7f0-6bf9-41e8-b126-9dd4509e3ee6', 'Outline Button', 'material-symbols:save', '/test/outline', 'outline', 'left', '');
INSERT INTO config.page_action_buttons VALUES ('e9c6327a-059e-4e70-a7b6-af376d0384b2', 'Ghost Button', 'material-symbols:download', '/test/ghost', 'ghost', 'left', '');
INSERT INTO config.page_action_buttons VALUES ('14464878-cdad-44ff-a781-a8446706ec52', 'Destructive Button', 'material-symbols:delete-forever', '/test/destructive', 'destructive', 'right', 'Are you sure you want to perform this destructive action? This is just a test.');
INSERT INTO config.page_action_buttons VALUES ('64914175-c59e-4337-a530-efb304483319', 'New User', 'material-symbols:person-add', '/admin/users/new', 'default', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('f18471cd-bd55-4d03-a914-5ba22e2b241b', 'New Asset', 'material-symbols:add-circle', '/assets/list/new', 'default', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('146be8e0-d43c-4e69-af6c-8792a37ac066', 'New Office', 'material-symbols:add-business', '/hr/offices/new', 'default', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('b37fe8cd-3135-40fa-93ec-3f98c677d416', 'New Transfer', 'material-symbols:add-circle', '/inventory/transfers/new', 'default', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('8dd86e75-5afc-4929-bfaa-1b1613cf4db5', 'New Purchase Order', 'material-symbols:note-add', '/procurement/orders/new', 'default', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('9a9b7b52-c5b9-456e-825e-69be8b4d5557', 'New Customer', 'material-symbols:person-add', '/sales/customers/new', 'default', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('70011703-9a8b-4803-b2df-f58f19c9afdd', 'New Warehouse', 'material-symbols:add-business', '/inventory/warehouses/new', 'default', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('a1107ca5-3b38-4f36-95a0-f5786966ae18', 'New Adjustment', 'material-symbols:tune', '/inventory/adjustments/new', 'default', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('0eb71ace-c6e2-4788-b91c-8e168bbe5744', 'New Customer', 'material-symbols:person-add', '/sales/customers/new', 'default', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('e2f67272-0037-4eb3-beba-1b8ee7c35298', 'New Order', 'material-symbols:add-shopping-cart', '/sales/orders/new', 'default', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('02ee40f0-c1d5-4fcd-a65f-6fb4550d5299', 'New Role', 'material-symbols:add-moderator', '/admin/roles/new', 'default', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('1dbe4633-46ca-4b35-b4a2-cbb598be1791', 'New Order', 'material-symbols:add-shopping-cart', '/sales/orders/new', 'default', 'right', '');
INSERT INTO config.page_action_buttons VALUES ('c35ed9f7-6b99-4377-8360-e0f903be7060', 'New Employee', 'material-symbols:person-add', '/hr/employees/new', 'default', 'right', '');


--
-- PostgreSQL database dump complete
--

\unrestrict xTl7qm0afGawHAHNllFqD3Ps4DRMH4gkp2KbAma7dOEe2HvfDE6AbG6xF7W5ISp

COMMIT;
