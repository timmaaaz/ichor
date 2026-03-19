BEGIN;
--
-- PostgreSQL database dump
--

\restrict VyQCQfBYJGsUkAFcns7XSvNjudfLjpsWTOCsZdYmyPig8eRGRLIrEOSysiZvN8k

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
-- Data for Name: currencies; Type: TABLE DATA; Schema: core; Owner: -
--

INSERT INTO core.currencies VALUES ('3d3e7fbf-6945-4d97-95e4-f8c50262dbb3', 'USD', 'US Dollar', '$', 'en-US', 2, true, 1, NULL, '2026-03-19 15:16:10.073592+00', NULL, '2026-03-19 15:16:10.073592+00');
INSERT INTO core.currencies VALUES ('59a53da4-1871-4351-85bf-6656b6729757', 'EUR', 'Euro', '€', 'en-EU', 2, true, 2, NULL, '2026-03-19 15:16:10.073592+00', NULL, '2026-03-19 15:16:10.073592+00');
INSERT INTO core.currencies VALUES ('60a6a348-9b4f-4bcf-99a2-6fda725135e0', 'GBP', 'British Pound', '£', 'en-GB', 2, true, 3, NULL, '2026-03-19 15:16:10.073592+00', NULL, '2026-03-19 15:16:10.073592+00');
INSERT INTO core.currencies VALUES ('0a03e4ef-6b34-46ea-a39f-3918b78794b1', 'CAD', 'Canadian Dollar', '$', 'en-CA', 2, true, 4, NULL, '2026-03-19 15:16:10.073592+00', NULL, '2026-03-19 15:16:10.073592+00');
INSERT INTO core.currencies VALUES ('764df0f6-a1d9-4eba-a0d0-109e6ea3bede', 'AUD', 'Australian Dollar', '$', 'en-AU', 2, true, 5, NULL, '2026-03-19 15:16:10.073592+00', NULL, '2026-03-19 15:16:10.073592+00');
INSERT INTO core.currencies VALUES ('53708cb5-172f-4481-84ee-5ca8c1a63299', 'JPY', 'Japanese Yen', '¥', 'ja-JP', 0, true, 6, NULL, '2026-03-19 15:16:10.073592+00', NULL, '2026-03-19 15:16:10.073592+00');
INSERT INTO core.currencies VALUES ('0ce3d5e3-e259-4f82-9928-50619579cdfb', 'CHF', 'Swiss Franc', 'CHF', 'de-CH', 2, true, 7, NULL, '2026-03-19 15:16:10.073592+00', NULL, '2026-03-19 15:16:10.073592+00');
INSERT INTO core.currencies VALUES ('327f5416-03c7-467d-b8ab-39ab1710cc8e', 'CNY', 'Chinese Yuan', '¥', 'zh-CN', 2, true, 8, NULL, '2026-03-19 15:16:10.073592+00', NULL, '2026-03-19 15:16:10.073592+00');
INSERT INTO core.currencies VALUES ('1b807d60-5f4c-434a-845c-3271d6bc6d9b', 'INR', 'Indian Rupee', '₹', 'en-IN', 2, true, 9, NULL, '2026-03-19 15:16:10.073592+00', NULL, '2026-03-19 15:16:10.073592+00');
INSERT INTO core.currencies VALUES ('b62d4c2d-1cf9-4570-8ef0-3fc5a6678ef9', 'MXN', 'Mexican Peso', '$', 'es-MX', 2, true, 10, NULL, '2026-03-19 15:16:10.073592+00', NULL, '2026-03-19 15:16:10.073592+00');
INSERT INTO core.currencies VALUES ('fcb138c3-933a-4aed-811c-4b8425cca861', 'TS0', 'Test Currency 0', '$0', 'en-US', 2, true, 100, NULL, '2026-03-19 15:16:51.276989+00', NULL, '2026-03-19 15:16:51.276989+00');
INSERT INTO core.currencies VALUES ('a40ad472-0d21-4baf-87dd-eb0eb1e4fffb', 'TS1', 'Test Currency 1', '$1', 'en-US', 2, true, 101, NULL, '2026-03-19 15:16:51.279514+00', NULL, '2026-03-19 15:16:51.279514+00');
INSERT INTO core.currencies VALUES ('66c3d72d-d83c-4b95-aad6-f47a44f415bf', 'TS2', 'Test Currency 2', '$2', 'en-US', 2, true, 102, NULL, '2026-03-19 15:16:51.282151+00', NULL, '2026-03-19 15:16:51.282151+00');
INSERT INTO core.currencies VALUES ('30b75a38-a3f5-4de0-90ae-a1333cf7ffcf', 'TS3', 'Test Currency 3', '$3', 'en-US', 2, true, 103, NULL, '2026-03-19 15:16:51.284717+00', NULL, '2026-03-19 15:16:51.284717+00');
INSERT INTO core.currencies VALUES ('c145daa3-94dc-4212-8465-9e30838e4c8a', 'TS4', 'Test Currency 4', '$4', 'en-US', 2, true, 104, NULL, '2026-03-19 15:16:51.287449+00', NULL, '2026-03-19 15:16:51.287449+00');


--
-- PostgreSQL database dump complete
--

\unrestrict VyQCQfBYJGsUkAFcns7XSvNjudfLjpsWTOCsZdYmyPig8eRGRLIrEOSysiZvN8k

COMMIT;
