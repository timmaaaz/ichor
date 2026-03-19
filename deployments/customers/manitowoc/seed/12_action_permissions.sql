BEGIN;
--
-- PostgreSQL database dump
--

\restrict bwBqhNZzE1yWWWdBawDWb1DspWhNf2yCcxt8u8TkbJw0VCAoPOXvY2fKZefVdDP

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
-- Data for Name: action_permissions; Type: TABLE DATA; Schema: workflow; Owner: -
--

INSERT INTO workflow.action_permissions VALUES ('1be4ec6e-66ab-4055-bbfa-334f2a02b70d', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'allocate_inventory', true, '{}', '2026-03-19 15:16:10.073592', '2026-03-19 15:16:10.073592');
INSERT INTO workflow.action_permissions VALUES ('e37b7601-c7a6-461b-845c-c131601e17fc', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'create_alert', true, '{}', '2026-03-19 15:16:10.073592', '2026-03-19 15:16:10.073592');
INSERT INTO workflow.action_permissions VALUES ('9522fd91-0400-4ae1-a03b-2d46902784fd', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'send_email', true, '{}', '2026-03-19 15:16:10.073592', '2026-03-19 15:16:10.073592');
INSERT INTO workflow.action_permissions VALUES ('c211e805-c560-4230-9a1a-92a5cb507d1b', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'send_notification', true, '{}', '2026-03-19 15:16:10.073592', '2026-03-19 15:16:10.073592');
INSERT INTO workflow.action_permissions VALUES ('88428f8b-ff28-43bd-9e70-29c7db7580bd', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'seek_approval', true, '{}', '2026-03-19 15:16:10.073592', '2026-03-19 15:16:10.073592');
INSERT INTO workflow.action_permissions VALUES ('79fbdba2-d9a4-40dc-ba07-83e0698c6d44', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'create_entity', true, '{}', '2026-03-19 15:16:10.073592', '2026-03-19 15:16:10.073592');
INSERT INTO workflow.action_permissions VALUES ('ee4d94c9-9662-496d-bc02-26e4664c933f', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'lookup_entity', true, '{}', '2026-03-19 15:16:10.073592', '2026-03-19 15:16:10.073592');
INSERT INTO workflow.action_permissions VALUES ('ec10dbce-4f3b-4de0-9626-7d7a63ed6d5c', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'transition_status', true, '{}', '2026-03-19 15:16:10.073592', '2026-03-19 15:16:10.073592');
INSERT INTO workflow.action_permissions VALUES ('addd455c-87be-436d-8a48-0a48d5625c29', '54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', 'log_audit_entry', true, '{}', '2026-03-19 15:16:10.073592', '2026-03-19 15:16:10.073592');


--
-- PostgreSQL database dump complete
--

\unrestrict bwBqhNZzE1yWWWdBawDWb1DspWhNf2yCcxt8u8TkbJw0VCAoPOXvY2fKZefVdDP

COMMIT;
