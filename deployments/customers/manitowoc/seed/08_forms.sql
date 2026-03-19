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
-- Data for Name: forms; Type: TABLE DATA; Schema: config; Owner: -
--

INSERT INTO config.forms VALUES ('3a1a985d-bd73-4525-bb61-21a88afb0f34', 'User Creation Form', false, false);
INSERT INTO config.forms VALUES ('eceeebbf-e92e-4179-a1b3-0fac591d0109', 'Asset Creation Form', false, false);
INSERT INTO config.forms VALUES ('60655c40-7c57-4178-8eb2-9ce38cf7cec8', 'User and Asset Creation Form', false, false);
INSERT INTO config.forms VALUES ('48fc5c17-ac2d-4611-8f4a-deef19ea0cf6', 'Full Customer Creation Form', false, false);
INSERT INTO config.forms VALUES ('7f0a2ace-8e9c-41a4-9d4f-e1b475ab604c', 'Full Supplier Creation Form', false, false);
INSERT INTO config.forms VALUES ('0efb19ee-f394-48da-84ad-e54c0b42a133', 'Full Sales Order Creation Form', false, false);
INSERT INTO config.forms VALUES ('7f6065b7-7265-435f-b7a0-1e55b7cf67b5', 'Full Purchase Order Creation Form', false, false);
INSERT INTO config.forms VALUES ('aa825c70-7642-43a8-9d2f-2ddd6ead469e', 'Role Creation Form', false, false);
INSERT INTO config.forms VALUES ('7d10f4d5-48b8-4fa5-9044-d1a1b220339f', 'Customer Creation Form', false, false);
INSERT INTO config.forms VALUES ('6a070255-1725-4236-9565-5d3c11873c24', 'Sales Order Creation Form', false, false);
INSERT INTO config.forms VALUES ('0ae4928b-d2b3-4727-85e8-fd7e780c3e84', 'Supplier Creation Form', false, false);
INSERT INTO config.forms VALUES ('3c686522-8120-472e-913a-849d0b4b4b85', 'Purchase Order Creation Form', false, false);
INSERT INTO config.forms VALUES ('206aca2d-3806-474a-947a-811ab83336d8', 'Warehouse Creation Form', false, false);
INSERT INTO config.forms VALUES ('e69fa15d-f38e-4390-b2ef-e469c074c57b', 'Inventory Adjustment Creation Form', false, false);
INSERT INTO config.forms VALUES ('252d20b0-5d45-47a6-9df3-32eed82671f4', 'Transfer Order Creation Form', false, false);
INSERT INTO config.forms VALUES ('6ccd397e-858b-46f3-a7d8-04f2ff837f23', 'Inventory Item Creation Form', false, false);
INSERT INTO config.forms VALUES ('79e44ab5-b3bf-458a-a26e-7df526a64524', 'Office Creation Form', false, false);
INSERT INTO config.forms VALUES ('7bfcbaf4-8664-466a-8050-a265dac37aec', 'Country Creation Form', false, false);
INSERT INTO config.forms VALUES ('7e7e67e9-7e10-4ca5-923b-866261cceb32', 'Region Creation Form', false, false);
INSERT INTO config.forms VALUES ('ff24f1f4-a6c4-4769-a3ca-dcb62689acfe', 'User Approval Status Creation Form', false, false);
INSERT INTO config.forms VALUES ('540747a1-2092-42a1-8d68-82fee578d90c', 'Asset Approval Status Creation Form', false, false);
INSERT INTO config.forms VALUES ('e30f1e8a-b92a-43f6-8b6c-e0afc0837d43', 'Asset Fulfillment Status Creation Form', false, false);
INSERT INTO config.forms VALUES ('800bf48b-aa42-4353-9c5d-de35472c6cac', 'Order Fulfillment Status Creation Form', false, false);
INSERT INTO config.forms VALUES ('c12b2442-4e44-420a-8f6f-0ebe89ad2498', 'Line Item Fulfillment Status Creation Form', false, false);
INSERT INTO config.forms VALUES ('bf83def8-af0e-4890-9bcc-c01722ab8691', 'Purchase Order Status Creation Form', false, false);
INSERT INTO config.forms VALUES ('b1a4b1c4-ec84-47fb-8993-4c3968007b90', 'Purchase Order Line Item Status Creation Form', false, false);
INSERT INTO config.forms VALUES ('8525a115-ea38-4090-851d-73065037faff', 'Asset Type Creation Form', false, false);
INSERT INTO config.forms VALUES ('bd14568e-ecf3-4f64-92b2-1658f959eff3', 'Asset Condition Creation Form', false, false);
INSERT INTO config.forms VALUES ('56a0c2c1-522b-46ec-b444-c5c4bc0043f7', 'Product Category Creation Form', false, false);
INSERT INTO config.forms VALUES ('5f59615b-0544-44de-a22a-0ba72308c266', 'City Creation Form', false, false);
INSERT INTO config.forms VALUES ('f6307039-d0ec-49f1-9b67-fe2b741bba25', 'Street Creation Form', false, false);
INSERT INTO config.forms VALUES ('86c797db-83d4-49bd-9a60-f8c89089256c', 'Contact Info Creation Form', false, false);
INSERT INTO config.forms VALUES ('8a8a3479-e789-4a57-8a9f-993f5f5be6d4', 'Title Creation Form', false, false);
INSERT INTO config.forms VALUES ('181990cb-eac9-4305-bd58-cedd594b0d0b', 'Product Creation Form', false, false);
INSERT INTO config.forms VALUES ('f3d44484-ad41-46d6-852d-6a04aa182df2', 'Inventory Location Creation Form', false, false);
INSERT INTO config.forms VALUES ('0e0e6bad-c7c7-4680-bad8-ab3663890ee5', 'Valid Asset Creation Form', false, false);
INSERT INTO config.forms VALUES ('a88b2f7a-ec84-4915-a464-9068c73f63e7', 'Supplier Product Creation Form', false, false);
INSERT INTO config.forms VALUES ('7df66fb2-e1e6-4d04-9260-3465ab2d4430', 'Sales Order Line Item Creation Form', false, false);
INSERT INTO config.forms VALUES ('be8ae182-b6fe-4744-b150-8400a689c88e', 'Purchase Order Line Item Creation Form', false, false);
INSERT INTO config.forms VALUES ('4237c243-079d-4d48-a218-0b197709bb5c', 'Home Creation Form', false, false);
INSERT INTO config.forms VALUES ('9ed3f8a1-9f8a-4765-81a4-28e4140005c0', 'Tag Creation Form', false, false);
INSERT INTO config.forms VALUES ('e9c787ae-2322-4ccd-805a-6deb3c141b5b', 'User Approval Comment Creation Form', false, false);
INSERT INTO config.forms VALUES ('da3e72e4-1549-4bd3-968e-aa19b199db45', 'Brand Creation Form', false, false);
INSERT INTO config.forms VALUES ('1cbe74de-c8b9-492a-b90a-8e2c2e52afd2', 'Zone Creation Form', false, false);
INSERT INTO config.forms VALUES ('14e8d097-4e01-415b-a0aa-44d3f0fa5109', 'User Asset Creation Form', false, false);
INSERT INTO config.forms VALUES ('2b4ec9b0-fadc-48d4-8300-f2bb42a627b5', 'Automation Rule Creation Form', false, false);
INSERT INTO config.forms VALUES ('a9d44f89-a4cb-4fea-a7a1-6fc6c7bfc336', 'Rule Action Creation Form', false, false);
INSERT INTO config.forms VALUES ('766d5a39-10b4-4cdc-8d43-9deb056805d2', 'Entity Creation Form', false, false);
INSERT INTO config.forms VALUES ('6cc1d3eb-9289-4835-99e4-a91eae38a71f', 'User Creation Form (Proper)', false, false);
INSERT INTO config.forms VALUES ('f6b34436-2467-4eca-b513-73b5493e657c', 'Physical Attribute Creation Form', false, false);
INSERT INTO config.forms VALUES ('3dabcd9d-d770-4da7-931e-3df90e9b9f0a', 'Product Cost Creation Form', false, false);
INSERT INTO config.forms VALUES ('09d1320d-ac7f-4f67-8175-0b1b14442d31', 'Cost History Creation Form', false, false);
INSERT INTO config.forms VALUES ('a4aaadad-6d6c-4550-a810-be118ce55ecd', 'Quality Metric Creation Form', false, false);
INSERT INTO config.forms VALUES ('685e7fb5-89c3-408e-8000-593d664e1d2b', 'Serial Number Creation Form', false, false);
INSERT INTO config.forms VALUES ('bc2e2b84-8ae8-4389-9470-a72378b4f3db', 'Lot Tracking Creation Form', false, false);
INSERT INTO config.forms VALUES ('7a8a9263-4bee-461f-943e-b96923fed805', 'Quality Inspection Creation Form', false, false);
INSERT INTO config.forms VALUES ('b69fd8e0-2ea5-40c0-8f10-df5e721b8376', 'Inventory Transaction Creation Form', false, false);


--
-- PostgreSQL database dump complete
--

\unrestrict sSw69w8gnFlHCtwmFDhZrSJtM1sSIliU8jxnXwOwHYzdkx18TXPstLYsK4Ewjxb

COMMIT;
