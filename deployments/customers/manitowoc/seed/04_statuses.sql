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
-- Data for Name: approval_status; Type: TABLE DATA; Schema: assets; Owner: -
--

INSERT INTO assets.approval_status VALUES ('b721d9d3-df81-46a4-8435-75d51c595fcc', 'ec451090-46ef-4ebc-8836-5481c2e1ebd2', 'SUCCESS', NULL, NULL, NULL);
INSERT INTO assets.approval_status VALUES ('441d5fd1-533d-42b3-bb64-8bdd39bab0f7', '1ead5362-31c6-42b3-be51-4c15c6bce7f4', 'ERROR', NULL, NULL, NULL);
INSERT INTO assets.approval_status VALUES ('d6e832cf-3033-4d33-b3e1-9779502a36b0', 'b9a5d0c8-7c5f-4fc5-bd08-bcfc2af8fd36', 'WAITING', NULL, NULL, NULL);
INSERT INTO assets.approval_status VALUES ('8e6e9692-e72d-49fc-ab2e-d2e1361b7df0', '258d3f18-3ebd-4924-bbf5-2daa0284e886', 'REJECTED', NULL, NULL, NULL);
INSERT INTO assets.approval_status VALUES ('d0115f72-9a58-42c6-a745-8ff0592f4396', '6162f903-3bc0-40ad-ad47-5f89cef572f2', 'IN_PROGRESS', NULL, NULL, NULL);
INSERT INTO assets.approval_status VALUES ('37bf478f-3f70-4632-bef3-dd9e4d114914', '6d86cb4a-d5de-4667-abf4-32a775c5474e', 'SUCCESS', NULL, NULL, NULL);
INSERT INTO assets.approval_status VALUES ('28e0480b-da1b-4fd0-9eed-91053bd2119b', '3cb6372c-7cdf-4e90-b2d0-23aa8a9cb8c5', 'ERROR', NULL, NULL, NULL);
INSERT INTO assets.approval_status VALUES ('8cfdead4-0a78-467d-9431-99c24f4663c4', '6ed67f11-7357-4f8e-8360-03a5515a8912', 'WAITING', NULL, NULL, NULL);
INSERT INTO assets.approval_status VALUES ('73bd1c36-493e-4413-becd-79f513abcefc', '0e86a7fa-a669-4710-b783-35672aeb5fe8', 'REJECTED', NULL, NULL, NULL);
INSERT INTO assets.approval_status VALUES ('0869c115-128c-4f01-9eb8-9af9d1dd34bf', 'c27ae605-cfab-451f-a264-927abad8c217', 'IN_PROGRESS', NULL, NULL, NULL);


--
-- Data for Name: asset_conditions; Type: TABLE DATA; Schema: assets; Owner: -
--

INSERT INTO assets.asset_conditions VALUES ('19c118a0-5379-44e6-9494-f59dbca82fcb', 'PERFECT', NULL);
INSERT INTO assets.asset_conditions VALUES ('cf6561c9-af6b-4e63-bde3-2b49377774ce', 'GOOD', NULL);
INSERT INTO assets.asset_conditions VALUES ('6e811fe8-6ac8-43f7-8ac8-779d861eba31', 'USED', NULL);
INSERT INTO assets.asset_conditions VALUES ('1bcba3a7-298d-4a3c-8d4d-5ac8f930e202', 'POOR', NULL);
INSERT INTO assets.asset_conditions VALUES ('67285801-9c05-42da-ab2a-e28910aa8581', 'END_OF_LIFE', NULL);
INSERT INTO assets.asset_conditions VALUES ('fd39db56-525c-4c21-b31d-df8683f33adb', 'AssetCondition9420', 'AssetCondition9420 Description');
INSERT INTO assets.asset_conditions VALUES ('95f3a46b-d7e5-4883-b59c-aada6ddfef54', 'AssetCondition9421', 'AssetCondition9421 Description');
INSERT INTO assets.asset_conditions VALUES ('04c2a404-c3dc-48e5-a9aa-504084456db6', 'AssetCondition9422', 'AssetCondition9422 Description');
INSERT INTO assets.asset_conditions VALUES ('66dc0c96-c434-4451-91f0-44d5d94d4673', 'AssetCondition9423', 'AssetCondition9423 Description');
INSERT INTO assets.asset_conditions VALUES ('2277be9e-82d7-4701-b74b-bd4a92506be2', 'AssetCondition9424', 'AssetCondition9424 Description');
INSERT INTO assets.asset_conditions VALUES ('22549859-035a-42bc-abd9-b369a34386a7', 'AssetCondition9425', 'AssetCondition9425 Description');
INSERT INTO assets.asset_conditions VALUES ('0b977bdb-b3a7-40ba-8c84-38a97541e3d4', 'AssetCondition9426', 'AssetCondition9426 Description');
INSERT INTO assets.asset_conditions VALUES ('09a299af-e5ea-42e5-9bed-f98fea1fe141', 'AssetCondition9427', 'AssetCondition9427 Description');


--
-- Data for Name: fulfillment_status; Type: TABLE DATA; Schema: assets; Owner: -
--

INSERT INTO assets.fulfillment_status VALUES ('73da3f34-d97f-4c80-ae84-108bedbe6bf1', '2365215c-d90c-4bb9-8495-f5b4a18c9736', 'SUCCESS', NULL, NULL, NULL);
INSERT INTO assets.fulfillment_status VALUES ('49e6b653-c8ff-4646-8f01-ee5d183c75f3', '2f2a5a4e-2390-41b1-b937-93a60a5c9d45', 'ERROR', NULL, NULL, NULL);
INSERT INTO assets.fulfillment_status VALUES ('f13846fa-f748-4a52-a6b0-9df4e04265a7', 'b84055bd-9884-4a56-9466-a2aa0d3d8e29', 'WAITING', NULL, NULL, NULL);
INSERT INTO assets.fulfillment_status VALUES ('96219e92-7120-4da2-b72e-b6cf63d75520', '769d1a21-d27b-422f-bece-e16b0555b3f3', 'REJECTED', NULL, NULL, NULL);
INSERT INTO assets.fulfillment_status VALUES ('f10f95c3-8f7c-40bf-b51c-e1e06e5a4477', 'e75c9b0c-a8ce-4497-afd7-2cb108c42b79', 'IN_PROGRESS', NULL, NULL, NULL);
INSERT INTO assets.fulfillment_status VALUES ('0c883e99-ee62-4a4b-a442-95e1d9b27e59', 'dbcf197e-bca2-4599-ba4b-1f7936979d23', 'SUCCESS', NULL, NULL, NULL);
INSERT INTO assets.fulfillment_status VALUES ('187c0b02-cca0-45a1-a8ac-f4747347ee58', '1b087881-3aeb-44ec-8475-52f74d22f1b1', 'ERROR', NULL, NULL, NULL);
INSERT INTO assets.fulfillment_status VALUES ('fab40c24-727b-4136-9d6b-dc9843823e02', '291c4dcc-95b5-40ff-90ab-46d76d918a61', 'WAITING', NULL, NULL, NULL);
INSERT INTO assets.fulfillment_status VALUES ('7d882d88-91eb-4b5f-94c8-65b4d782efca', '9c25473f-cdcc-4653-9ac3-793630bb5cbd', 'REJECTED', NULL, NULL, NULL);
INSERT INTO assets.fulfillment_status VALUES ('4c9182e6-d6c0-4a3e-8d0d-0d24f86001c8', '5e4ce3eb-c6c9-406f-b5b7-f041c6611969', 'IN_PROGRESS', NULL, NULL, NULL);


--
-- Data for Name: payment_terms; Type: TABLE DATA; Schema: core; Owner: -
--

INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000001', 'Net 30', 'Payment due within 30 days of invoice date');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000002', 'Net 60', 'Payment due within 60 days of invoice date');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000003', 'Due on Receipt', 'Payment due immediately upon receipt of invoice');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000004', 'Net 15', 'Payment due within 15 days of invoice date');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000005', 'Net 45', 'Payment due within 45 days of invoice date');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000006', 'Net 7', 'Payment due within 7 days of invoice date');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000007', 'Net 10', 'Payment due within 10 days of invoice date');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000008', 'Net 21', 'Payment due within 21 days of invoice date');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000009', 'Net 90', 'Payment due within 90 days of invoice date');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000010', 'Prepaid', 'Full payment required before order fulfillment');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000011', 'COD', 'Cash on Delivery - payment due at time of delivery');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000012', 'CIA', 'Cash in Advance - full payment before shipping');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000013', '50% Deposit', '50% payment due upfront, balance on delivery');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000014', 'EOM', 'Payment due at end of month');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000015', 'MFI', 'Month Following Invoice - due end of next month');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000016', '15 MFI', '15 days after month following invoice');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000017', 'Open Account', 'Standard credit account with flexible terms');
INSERT INTO core.payment_terms VALUES ('a0000000-0000-4000-8000-000000000018', 'Letter of Credit', 'Payment secured by letter of credit');


--
-- Data for Name: user_approval_status; Type: TABLE DATA; Schema: hr; Owner: -
--

INSERT INTO hr.user_approval_status VALUES ('89173300-3f4e-4606-872c-f34914bbee19', 'cf92611d-52a9-46f6-bd3d-dcc18d2e0a8e', 'PENDING', NULL, NULL, NULL);
INSERT INTO hr.user_approval_status VALUES ('0394acac-ace4-4e8f-b64e-68625b0af14a', 'b031d5f0-23bc-4026-8eb7-d4c3fd0b0f7a', 'APPROVED', NULL, NULL, NULL);
INSERT INTO hr.user_approval_status VALUES ('7b901e2e-3f33-40c1-9201-b4e8b1718b4b', 'ef9c7d4d-1ff6-4009-bd92-6aafb3944c4f', 'DENIED', NULL, NULL, NULL);
INSERT INTO hr.user_approval_status VALUES ('132a2572-b7a0-4b56-a165-55e1c244c3e2', '8370b848-736c-4c55-bb9b-aae788049f9f', 'UNDER REVIEW', NULL, NULL, NULL);


--
-- PostgreSQL database dump complete
--

\unrestrict o4eDhNdUea9vqLeRMVHnpYyeIIyUC1FYoXjtHu2Sdm9MVs2f0S6THFET4ABcbHH

COMMIT;
