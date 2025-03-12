CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

INSERT INTO countries (country_id, number, name, alpha_2, alpha_3) VALUES
    (uuid_generate_v4(),1,'Andorra','AD','AND'),
    (uuid_generate_v4(),2,'United Arab Emirates','AE','ARE'),
    (uuid_generate_v4(),3,'Afghanistan','AF','AFG'),
    (uuid_generate_v4(),4,'Antigua and Barbuda','AG','ATG'),
    (uuid_generate_v4(),5,'Anguilla','AI','AIA'),
    (uuid_generate_v4(),6,'Albania','AL','ALB'),
    (uuid_generate_v4(),7,'Armenia','AM','ARM'),
    (uuid_generate_v4(),8,'Netherlands Antilles','AN','ANT'),
    (uuid_generate_v4(),9,'Angola','AO','AGO'),
    (uuid_generate_v4(),10,'Antarctica','AQ','ATA'),
    (uuid_generate_v4(),11,'Argentina','AR','ARG'),
    (uuid_generate_v4(),12,'American Samoa','AS','ASM'),
    (uuid_generate_v4(),13,'Austria','AT','AUT'),
    (uuid_generate_v4(),14,'Australia','AU','AUS'),
    (uuid_generate_v4(),15,'Aruba','AW','ABW'),
    (uuid_generate_v4(),16,'Aland Islands','AX','ALA'),
    (uuid_generate_v4(),17,'Azerbaijan','AZ','AZE'),
    (uuid_generate_v4(),18,'Bosnia and Herzegovina','BA','BIH'),
    (uuid_generate_v4(),19,'Barbados','BB','BRB'),
    (uuid_generate_v4(),20,'Bangladesh','BD','BGD'),
    (uuid_generate_v4(),21,'Belgium','BE','BEL'),
    (uuid_generate_v4(),22,'Burkina Faso','BF','BFA'),
    (uuid_generate_v4(),23,'Bulgaria','BG','BGR'),
    (uuid_generate_v4(),24,'Bahrain','BH','BHR'),
    (uuid_generate_v4(),25,'Burundi','BI','BDI'),
    (uuid_generate_v4(),26,'Benin','BJ','BEN'),
    (uuid_generate_v4(),27,'Saint Barthelemy','BL','BLM'),
    (uuid_generate_v4(),28,'Bermuda','BM','BMU'),
    (uuid_generate_v4(),29,'Brunei Darussalam','BN','BRN'),
    (uuid_generate_v4(),30,'Bolivia','BO','BOL'),
    (uuid_generate_v4(),31,'Brazil','BR','BRA'),
    (uuid_generate_v4(),32,'Bahamas','BS','BHS'),
    (uuid_generate_v4(),33,'Bhutan','BT','BTN'),
    (uuid_generate_v4(),34,'Bouvet Island','BV','BVT'),
    (uuid_generate_v4(),35,'Botswana','BW','BWA'),
    (uuid_generate_v4(),36,'Belarus','BY','BLR'),
    (uuid_generate_v4(),37,'Belize','BZ','BLZ'),
    (uuid_generate_v4(),38,'Canada','CA','CAN'),
    (uuid_generate_v4(),39,'Cocos (Keeling) Islands','CC','CCK'),
    (uuid_generate_v4(),40,'Congo, The Democratic Republic Of The','CD','COD'),
    (uuid_generate_v4(),41,'Central African Republic','CF','CAF'),
    (uuid_generate_v4(),42,'Congo','CG','COG'),
    (uuid_generate_v4(),43,'Switzerland','CH','CHE'),
    (uuid_generate_v4(),44,'Cote D''Ivoire','CI','CIV'),
    (uuid_generate_v4(),45,'Cook Islands','CK','COK'),
    (uuid_generate_v4(),46,'Chile','CL','CHL'),
    (uuid_generate_v4(),47,'Cameroon','CM','CMR'),
    (uuid_generate_v4(),48,'China','CN','CHN'),
    (uuid_generate_v4(),49,'Colombia','CO','COL'),
    (uuid_generate_v4(),50,'Costa Rica','CR','CRI'),
    (uuid_generate_v4(),51,'Cuba','CU','CUB'),
    (uuid_generate_v4(),52,'Cape Verde','CV','CPV'),
    (uuid_generate_v4(),53,'Christmas Island','CX','CXR'),
    (uuid_generate_v4(),54,'Cyprus','CY','CYP'),
    (uuid_generate_v4(),55,'Czech Republic','CZ','CZE'),
    (uuid_generate_v4(),56,'Germany','DE','DEU'),
    (uuid_generate_v4(),57,'Djibouti','DJ','DJI'),
    (uuid_generate_v4(),58,'Denmark','DK','DNK'),
    (uuid_generate_v4(),59,'Dominica','DM','DMA'),
    (uuid_generate_v4(),60,'Dominican Republic','DO','DOM'),
    (uuid_generate_v4(),61,'Algeria','DZ','DZA'),
    (uuid_generate_v4(),62,'Ecuador','EC','ECU'),
    (uuid_generate_v4(),63,'Estonia','EE','EST'),
    (uuid_generate_v4(),64,'Egypt','EG','EGY'),
    (uuid_generate_v4(),65,'Western Sahara','EH','ESH'),
    (uuid_generate_v4(),66,'Eritrea','ER','ERI'),
    (uuid_generate_v4(),67,'Spain','ES','ESP'),
    (uuid_generate_v4(),68,'Ethiopia','ET','ETH'),
    (uuid_generate_v4(),69,'Finland','FI','FIN'),
    (uuid_generate_v4(),70,'Fiji','FJ','FJI'),
    (uuid_generate_v4(),71,'Falkland Islands (Malvinas)','FK','FLK'),
    (uuid_generate_v4(),72,'Micronesia','FM','FSM'),
    (uuid_generate_v4(),73,'Faroe Islands','FO','FRO'),
    (uuid_generate_v4(),74,'France','FR','FRA'),
    (uuid_generate_v4(),75,'Gabon','GA','GAB'),
    (uuid_generate_v4(),76,'United Kingdom','GB','GBR'),
    (uuid_generate_v4(),77,'Grenada','GD','GRD'),
    (uuid_generate_v4(),78,'Georgia','GE','GEO'),
    (uuid_generate_v4(),79,'French Guiana','GF','GUF'),
    (uuid_generate_v4(),80,'Guernsey','GG','GGY'),
    (uuid_generate_v4(),81,'Ghana','GH','GHA'),
    (uuid_generate_v4(),82,'Gibraltar','GI','GIB'),
    (uuid_generate_v4(),83,'Greenland','GL','GRL'),
    (uuid_generate_v4(),84,'Gambia','GM','GMB'),
    (uuid_generate_v4(),85,'Guinea','GN','GIN'),
    (uuid_generate_v4(),86,'Guadeloupe','GP','GLP'),
    (uuid_generate_v4(),87,'Equatorial Guinea','GQ','GNQ'),
    (uuid_generate_v4(),88,'Greece','GR','GRC'),
    (uuid_generate_v4(),89,'South Georgia and the South Sandwich Islands','GS','SGS'),
    (uuid_generate_v4(),90,'Guatemala','GT','GTM'),
    (uuid_generate_v4(),91,'Guam','GU','GUM'),
    (uuid_generate_v4(),92,'Guinea-Bissau','GW','GNB'),
    (uuid_generate_v4(),93,'Guyana','GY','GUY'),
    (uuid_generate_v4(),94,'Hong Kong','HK','HKG'),
    (uuid_generate_v4(),95,'Heard Island and Mcdonald Islands','HM','HMD'),
    (uuid_generate_v4(),96,'Honduras','HN','HND'),
    (uuid_generate_v4(),97,'Croatia','HR','HRV'),
    (uuid_generate_v4(),98,'Haiti','HT','HTI'),
    (uuid_generate_v4(),99,'Hungary','HU','HUN'),
    (uuid_generate_v4(),100,'Indonesia','ID','IDN'),
    (uuid_generate_v4(),101,'Ireland','IE','IRL'),
    (uuid_generate_v4(),102,'Israel','IL','ISR'),
    (uuid_generate_v4(),103,'Isle of Man','IM','IMN'),
    (uuid_generate_v4(),104,'British Indian Ocean Territory','IO','IOT'),
    (uuid_generate_v4(),105,'Iraq','IQ','IRQ'),
    (uuid_generate_v4(),106,'Iran','IR','IRN'),
    (uuid_generate_v4(),107,'Iceland','IS','ISL'),
    (uuid_generate_v4(),108,'Italy','IT','ITA'),
    (uuid_generate_v4(),109,'Jersey','JE','JEY'),
    (uuid_generate_v4(),110,'Jamaica','JM','JAM'),
    (uuid_generate_v4(),111,'Jordan','JO','JOR'),
    (uuid_generate_v4(),112,'Japan','JP','JPN'),
    (uuid_generate_v4(),113,'Kenya','KE','KEN'),
    (uuid_generate_v4(),114,'Kyrgyzstan','KG','KGZ'),
    (uuid_generate_v4(),115,'Cambodia','KH','KHM'),
    (uuid_generate_v4(),116,'Kiribati','KI','KIR'),
    (uuid_generate_v4(),117,'Comoros','KM','COM'),
    (uuid_generate_v4(),118,'Saint Kitts and Nevis','KN','KNA'),
    (uuid_generate_v4(),119,'North Korea','KP','PRK'),
    (uuid_generate_v4(),120,'South Korea','KR','KOR'),
    (uuid_generate_v4(),121,'Kuwait','KW','KWT'),
    (uuid_generate_v4(),122,'Cayman Islands','KY','CYM'),
    (uuid_generate_v4(),123,'Kazakhstan','KZ','KAZ'),
    (uuid_generate_v4(),124,'Lao People''s Democratic Republic','LA','LAO'),
    (uuid_generate_v4(),125,'Lebanon','LB','LBN'),
    (uuid_generate_v4(),126,'Saint Lucia','LC','LCA'),
    (uuid_generate_v4(),127,'Liechtenstein','LI','LIE'),
    (uuid_generate_v4(),128,'Sri Lanka','LK','LKA'),
    (uuid_generate_v4(),129,'Liberia','LR','LBR'),
    (uuid_generate_v4(),130,'Lesotho','LS','LSO'),
    (uuid_generate_v4(),131,'Lithuania','LT','LTU'),
    (uuid_generate_v4(),132,'Luxembourg','LU','LUX'),
    (uuid_generate_v4(),133,'Latvia','LV','LVA'),
    (uuid_generate_v4(),134,'Libya','LY','LBY'),
    (uuid_generate_v4(),135,'Morocco','MA','MAR'),
    (uuid_generate_v4(),136,'Monaco','MC','MCO'),
    (uuid_generate_v4(),137,'Moldova','MD','MDA'),
    (uuid_generate_v4(),138,'Montenegro','ME','MNE'),
    (uuid_generate_v4(),139,'Saint Martin','MF','MAF'),
    (uuid_generate_v4(),140,'Madagascar','MG','MDG'),
    (uuid_generate_v4(),141,'Marshall Islands','MH','MHL'),
    (uuid_generate_v4(),142,'Macedonia','MK','MKD'),
    (uuid_generate_v4(),143,'Mali','ML','MLI'),
    (uuid_generate_v4(),144,'Myanmar','MM','MMR'),
    (uuid_generate_v4(),145,'Mongolia','MN','MNG'),
    (uuid_generate_v4(),146,'Macao','MO','MAC'),
    (uuid_generate_v4(),147,'Northern Mariana Islands','MP','MNP'),
    (uuid_generate_v4(),148,'Martinique','MQ','MTQ'),
    (uuid_generate_v4(),149,'Mauritania','MR','MRT'),
    (uuid_generate_v4(),150,'Montserrat','MS','MSR'),
    (uuid_generate_v4(),151,'Malta','MT','MLT'),
    (uuid_generate_v4(),152,'Mauritius','MU','MUS'),
    (uuid_generate_v4(),153,'Maldives','MV','MDV'),
    (uuid_generate_v4(),154,'Malawi','MW','MWI'),
    (uuid_generate_v4(),155,'Mexico','MX','MEX'),
    (uuid_generate_v4(),156,'Malaysia','MY','MYS'),
    (uuid_generate_v4(),157,'Mozambique','MZ','MOZ'),
    (uuid_generate_v4(),158,'Namibia','NA','NAM'),
    (uuid_generate_v4(),159,'New Caledonia','NC','NCL'),
    (uuid_generate_v4(),160,'Niger','NE','NER'),
    (uuid_generate_v4(),161,'Norfolk Island','NF','NFK'),
    (uuid_generate_v4(),162,'Nigeria','NG','NGA'),
    (uuid_generate_v4(),163,'Nicaragua','NI','NIC'),
    (uuid_generate_v4(),164,'Netherlands','NL','NLD'),
    (uuid_generate_v4(),165,'Norway','NO','NOR'),
    (uuid_generate_v4(),166,'Nepal','NP','NPL'),
    (uuid_generate_v4(),167,'Nauru','NR','NRU'),
    (uuid_generate_v4(),168,'Niue','NU','NIU'),
    (uuid_generate_v4(),169,'New Zealand','NZ','NZL'),
    (uuid_generate_v4(),170,'Oman','OM','OMN'),
    (uuid_generate_v4(),171,'Panama','PA','PAN'),
    (uuid_generate_v4(),172,'Peru','PE','PER'),
    (uuid_generate_v4(),173,'French Polynesia','PF','PYF'),
    (uuid_generate_v4(),174,'Papua New Guinea','PG','PNG'),
    (uuid_generate_v4(),175,'Philippines','PH','PHL'),
    (uuid_generate_v4(),176,'Pakistan','PK','PAK'),
    (uuid_generate_v4(),177,'Poland','PL','POL'),
    (uuid_generate_v4(),178,'Saint Pierre and Miquelon','PM','SPM'),
    (uuid_generate_v4(),179,'Pitcairn','PN','PCN'),
    (uuid_generate_v4(),180,'Puerto Rico','PR','PRI'),
    (uuid_generate_v4(),181,'Palestine','PS','PSE'),
    (uuid_generate_v4(),182,'Portugal','PT','PRT'),
    (uuid_generate_v4(),183,'Palau','PW','PLW'),
    (uuid_generate_v4(),184,'Paraguay','PY','PRY'),
    (uuid_generate_v4(),185,'Qatar','QA','QAT'),
    (uuid_generate_v4(),186,'Reunion','RE','REU'),
    (uuid_generate_v4(),187,'Romania','RO','ROU'),
    (uuid_generate_v4(),188,'Serbia','RS','SRB'),
    (uuid_generate_v4(),189,'Russia','RU','RUS'),
    (uuid_generate_v4(),190,'Rwanda','RW','RWA'),
    (uuid_generate_v4(),191,'Saudi Arabia','SA','SAU'),
    (uuid_generate_v4(),192,'Solomon Islands','SB','SLB'),
    (uuid_generate_v4(),193,'Seychelles','SC','SYC'),
    (uuid_generate_v4(),194,'Sudan','SD','SDN'),
    (uuid_generate_v4(),195,'Sweden','SE','SWE'),
    (uuid_generate_v4(),196,'Singapore','SG','SGP'),
    (uuid_generate_v4(),197,'Saint Helena','SH','SHN'),
    (uuid_generate_v4(),198,'Slovenia','SI','SVN'),
    (uuid_generate_v4(),199,'Svalbard and Jan Mayen','SJ','SJM'),
    (uuid_generate_v4(),200,'Slovakia','SK','SVK'),
    (uuid_generate_v4(),201,'Sierra Leone','SL','SLE'),
    (uuid_generate_v4(),202,'San Marino','SM','SMR'),
    (uuid_generate_v4(),203,'Senegal','SN','SEN'),
    (uuid_generate_v4(),204,'Somalia','SO','SOM'),
    (uuid_generate_v4(),205,'Suriname','SR','SUR'),
    (uuid_generate_v4(),206,'Sao Tome and Principe','ST','STP'),
    (uuid_generate_v4(),207,'El Salvador','SV','SLV'),
    (uuid_generate_v4(),208,'Syrian Arab Republic','SY','SYR'),
    (uuid_generate_v4(),209,'Swaziland','SZ','SWZ'),
    (uuid_generate_v4(),210,'Turks and Caicos Islands','TC','TCA'),
    (uuid_generate_v4(),211,'Chad','TD','TCD'),
    (uuid_generate_v4(),212,'French Southern Territories','TF','ATF'),
    (uuid_generate_v4(),213,'Togo','TG','TGO'),
    (uuid_generate_v4(),214,'Thailand','TH','THA'),
    (uuid_generate_v4(),215,'Tajikistan','TJ','TJK'),
    (uuid_generate_v4(),216,'Tokelau','TK','TKL'),
    (uuid_generate_v4(),217,'Timor-Leste','TL','TLS'),
    (uuid_generate_v4(),218,'Turkmenistan','TM','TKM'),
    (uuid_generate_v4(),219,'Tunisia','TN','TUN'),
    (uuid_generate_v4(),220,'Tonga','TO','TON'),
    (uuid_generate_v4(),221,'Turkey','TR','TUR'),
    (uuid_generate_v4(),222,'Trinidad and Tobago','TT','TTO'),
    (uuid_generate_v4(),223,'Tuvalu','TV','TUV'),
    (uuid_generate_v4(),224,'Taiwan','TW','TWN'),
    (uuid_generate_v4(),225,'Tanzania, United Republic Of','TZ','TZA'),
    (uuid_generate_v4(),226,'Ukraine','UA','UKR'),
    (uuid_generate_v4(),227,'Uganda','UG','UGA'),
    (uuid_generate_v4(),228,'United States Minor Outlying Islands','UM','UMI'),
    (uuid_generate_v4(),229,'United States','US','USA'),
    (uuid_generate_v4(),230,'Uruguay','UY','URY'),
    (uuid_generate_v4(),231,'Uzbekistan','UZ','UZB'),
    (uuid_generate_v4(),232,'Holy See (Vatican City State)','VA','VAT'),
    (uuid_generate_v4(),233,'Saint Vincent and the Grenadines','VC','VCT'),
    (uuid_generate_v4(),234,'Venezuela','VE','VEN'),
    (uuid_generate_v4(),235,'British Virgin Islands','VG','VGB'),
    (uuid_generate_v4(),236,'U.S. Virgin Islands','VI','VIR'),
    (uuid_generate_v4(),237,'Viet Nam','VN','VNM'),
    (uuid_generate_v4(),238,'Vanuatu','VU','VUT'),
    (uuid_generate_v4(),239,'Wallis and Futuna','WF','WLF'),
    (uuid_generate_v4(),240,'Samoa','WS','WSM'),
    (uuid_generate_v4(),241,'Yemen','YE','YEM'),
    (uuid_generate_v4(),242,'Mayotte','YT','MYT'),
    (uuid_generate_v4(),243,'South Africa','ZA','ZAF'),
    (uuid_generate_v4(),244,'Zambia','ZM','ZMB'),
    (uuid_generate_v4(),245,'Zimbabwe','ZW','ZWE'),
    (uuid_generate_v4(),246,'India','IN','IND'),
    (uuid_generate_v4(),247,'Kosovo','XK','XXK')
;
-- Version: 1.02
-- Description: Insert data into regions table
INSERT INTO regions (region_id, country_id, name, code) VALUES
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Alabama', 'AL'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Alaska', 'AK'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Arizona', 'AZ'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Arkansas', 'AR'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'California', 'CA'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Colorado', 'CO'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Connecticut', 'CT'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Delaware', 'DE'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Florida', 'FL'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Georgia', 'GA'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Hawaii', 'HI'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Idaho', 'ID'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Illinois', 'IL'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Indiana', 'IN'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Iowa', 'IA'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Kansas', 'KS'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Kentucky', 'KY'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Louisiana', 'LA'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Maine', 'ME'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Maryland', 'MD'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Massachusetts', 'MA'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Michigan', 'MI'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Minnesota', 'MN'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Mississippi', 'MS'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Missouri', 'MO'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Montana', 'MT'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Nebraska', 'NE'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Nevada', 'NV'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'New Hampshire', 'NH'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'New Jersey', 'NJ'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'New Mexico', 'NM'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'New York', 'NY'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'North Carolina', 'NC'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'North Dakota', 'ND'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Ohio', 'OH'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Oklahoma', 'OK'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Oregon', 'OR'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Pennsylvania', 'PA'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Rhode Island', 'RI'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'South Carolina', 'SC'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'South Dakota', 'SD'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Tennessee', 'TN'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Texas', 'TX'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Utah', 'UT'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Vermont', 'VT'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Virginia', 'VA'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Washington', 'WA'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'West Virginia', 'WV'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Wisconsin', 'WI'),
    (uuid_generate_v4(), (SELECT country_id FROM countries WHERE alpha_2 = 'US'), 'Wyoming', 'WY')
;

INSERT INTO roles (role_id, name, description)
VALUES 
    (gen_random_uuid(), 'ADMIN', 'System Administrator with full access'),
    (gen_random_uuid(), 'EMPLOYEE', 'Regular employee with standard access'),
    (gen_random_uuid(), 'FINANCE_ADMIN', 'Finance Department Administrator'),
    (gen_random_uuid(), 'HR_ADMIN', 'Human Resources Administrator'),
    (gen_random_uuid(), 'MANAGER', 'Department manager with extended privileges'),
    (gen_random_uuid(), 'READONLY', 'Read-only access to specific resources'),
    (gen_random_uuid(), 'TEMP_ADMIN', 'Temporary Administrator access');

INSERT INTO user_approval_status (user_approval_status_id, name, icon_id) VALUES
    ('89173300-3f4e-4606-872c-f34914bbee19', 'PENDING', uuid_generate_v4()),
    ('0394acac-ace4-4e8f-b64e-68625b0af14a', 'APPROVED', uuid_generate_v4()),
    ('7b901e2e-3f33-40c1-9201-b4e8b1718b4b', 'DENIED', uuid_generate_v4()),
    ('132a2572-b7a0-4b56-a165-55e1c244c3e2', 'UNDER REVIEW', uuid_generate_v4());

-- Insert 20 users with the provided structure
INSERT INTO users (
    user_id, username, first_name, last_name, birthday, email, title_id, work_phone_id, cell_phone_id, 
    requested_by, approved_by, roles, system_roles, password_hash, office_id, enabled, 
    date_requested, date_approved, date_created, date_updated, user_approval_status
) VALUES
    ('5cf37266-3473-4006-984f-9325122678b7', 'admin_gopher', 'Admin', 'Gopher', NULL, 'admin@example.com', NULL, NULL, NULL, NULL, NULL, '{ADMIN}', '{}', '$2a$10$1ggfMVZV6Js0ybvJufLRUOWHS5f6KneuP0XwwHpJ8L8ipdry9f2/a', NULL, true, NULL, NULL, '2019-03-24 00:00:00', '2019-03-24 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    ('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'user_gopher', 'User', 'Gopher', NULL, 'user@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2019-03-24 00:00:00', '2019-03-24 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'manager1', 'Kelly', 'Manager', NULL, 'manager1@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-01-15 00:00:00', '2023-01-15 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'manager2', 'Sam', 'Director', NULL, 'manager2@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-01-20 00:00:00', '2023-01-20 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'finance_admin', 'Morgan', 'Accountant', NULL, 'finance_admin@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-02-01 00:00:00', '2023-02-01 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'hr_admin', 'Pat', 'People', NULL, 'hr_admin@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-02-05 00:00:00', '2023-02-05 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'employee1', 'Alex', 'Worker', NULL, 'employee1@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-03-01 00:00:00', '2023-03-01 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'employee2', 'Taylor', 'Staff', NULL, 'employee2@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-03-05 00:00:00', '2023-03-05 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'employee3', 'Jordan', 'Associate', NULL, 'employee3@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-03-10 00:00:00', '2023-03-10 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'employee4', 'Casey', 'Analyst', NULL, 'employee4@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-03-15 00:00:00', '2023-03-15 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'readonly', 'Riley', 'Viewer', NULL, 'readonly@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-04-01 00:00:00', '2023-04-01 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'temp_admin', 'Dana', 'Temporary', NULL, 'temp_admin@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-04-05 00:00:00', '2023-04-05 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'sales_east', 'Quinn', 'Eastern', NULL, 'sales_east@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-04-10 00:00:00', '2023-04-10 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'sales_west', 'Avery', 'Western', NULL, 'sales_west@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-04-15 00:00:00', '2023-04-15 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'it_systems', 'Morgan', 'Tech', NULL, 'it_systems@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-05-01 00:00:00', '2023-05-01 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'it_dev', 'Jamie', 'Developer', NULL, 'it_dev@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-05-05 00:00:00', '2023-05-05 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'accounting', 'Skyler', 'Books', NULL, 'accounting@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-05-10 00:00:00', '2023-05-10 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'payroll', 'Drew', 'Payment', NULL, 'payroll@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-05-15 00:00:00', '2023-05-15 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'recruitment', 'Harley', 'Hiring', NULL, 'recruitment@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-06-01 00:00:00', '2023-06-01 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a'),
    (gen_random_uuid(), 'benefits', 'Charlie', 'Benefits', NULL, 'benefits@example.com', NULL, NULL, NULL, NULL, NULL, '{USER}', '{}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', NULL, true, NULL, NULL, '2023-06-05 00:00:00', '2023-06-05 00:00:00', '0394acac-ace4-4e8f-b64e-68625b0af14a')
;

-- Insert restricted columns
INSERT INTO restricted_columns (restricted_column_id, table_name, column_name)
VALUES
    (gen_random_uuid(), 'roles', 'description'),
    (gen_random_uuid(), 'users', 'email'),
    (gen_random_uuid(), 'user_organizations', 'is_unit_manager'),
    (gen_random_uuid(), 'user_organizations', 'role_id'),
    (gen_random_uuid(), 'users', 'user_id'),
    (gen_random_uuid(), 'users', 'username');

-- Create variable to store role IDs for later use
-- Insert roles first to ensure they exist before referencing them

-- Create variable to store role IDs for later use
DO $$
DECLARE
    admin_role_id UUID;
    employee_role_id UUID;
    finance_admin_role_id UUID;
    hr_admin_role_id UUID;
    manager_role_id UUID;
    readonly_role_id UUID;
    temp_admin_role_id UUID;
    
    admin_user_id UUID;
    user_user_id UUID;
    manager1_user_id UUID;
    manager2_user_id UUID;
    finance_admin_user_id UUID;
    hr_admin_user_id UUID;
    employee1_user_id UUID;
    employee2_user_id UUID;
    employee3_user_id UUID;
    employee4_user_id UUID;
    readonly_user_id UUID;
    temp_admin_user_id UUID;
    sales_east_user_id UUID;
    sales_west_user_id UUID;
    it_systems_user_id UUID;
    it_dev_user_id UUID;
    accounting_user_id UUID;
    payroll_user_id UUID;
    recruitment_user_id UUID;
    benefits_user_id UUID;
    
    hq_id UUID;
    finance_dept_id UUID;
    hr_dept_id UUID;
    sales_dept_id UUID;
    it_dept_id UUID;
    accounting_team_id UUID;
    payroll_team_id UUID;
    recruitment_team_id UUID;
    benefits_team_id UUID;
    east_region_id UUID;
    west_region_id UUID;
    systems_team_id UUID;
    dev_team_id UUID;
    northeast_branch_id UUID;
    southeast_branch_id UUID;
    northwest_branch_id UUID;
    southwest_branch_id UUID;

BEGIN
    -- Get role IDs - using LIMIT 1 to ensure only one row is returned
    SELECT role_id INTO admin_role_id FROM roles WHERE name = 'ADMIN' LIMIT 1;
    SELECT role_id INTO employee_role_id FROM roles WHERE name = 'EMPLOYEE' LIMIT 1;
    SELECT role_id INTO finance_admin_role_id FROM roles WHERE name = 'FINANCE_ADMIN' LIMIT 1;
    SELECT role_id INTO hr_admin_role_id FROM roles WHERE name = 'HR_ADMIN' LIMIT 1;
    SELECT role_id INTO manager_role_id FROM roles WHERE name = 'MANAGER' LIMIT 1;
    SELECT role_id INTO readonly_role_id FROM roles WHERE name = 'READONLY' LIMIT 1;
    SELECT role_id INTO temp_admin_role_id FROM roles WHERE name = 'TEMP_ADMIN' LIMIT 1;
    
    -- Get user IDs - using LIMIT 1 to ensure only one row is returned
    SELECT user_id INTO admin_user_id FROM users WHERE username = 'admin_gopher' LIMIT 1;
    SELECT user_id INTO user_user_id FROM users WHERE username = 'user_gopher' LIMIT 1;
    SELECT user_id INTO manager1_user_id FROM users WHERE username = 'manager1' LIMIT 1;
    SELECT user_id INTO manager2_user_id FROM users WHERE username = 'manager2' LIMIT 1;
    SELECT user_id INTO finance_admin_user_id FROM users WHERE username = 'finance_admin' LIMIT 1;
    SELECT user_id INTO hr_admin_user_id FROM users WHERE username = 'hr_admin' LIMIT 1;
    SELECT user_id INTO employee1_user_id FROM users WHERE username = 'employee1' LIMIT 1;
    SELECT user_id INTO employee2_user_id FROM users WHERE username = 'employee2' LIMIT 1;
    SELECT user_id INTO employee3_user_id FROM users WHERE username = 'employee3' LIMIT 1;
    SELECT user_id INTO employee4_user_id FROM users WHERE username = 'employee4' LIMIT 1;
    SELECT user_id INTO readonly_user_id FROM users WHERE username = 'readonly' LIMIT 1;
    SELECT user_id INTO temp_admin_user_id FROM users WHERE username = 'temp_admin' LIMIT 1;
    SELECT user_id INTO sales_east_user_id FROM users WHERE username = 'sales_east' LIMIT 1;
    SELECT user_id INTO sales_west_user_id FROM users WHERE username = 'sales_west' LIMIT 1;
    SELECT user_id INTO it_systems_user_id FROM users WHERE username = 'it_systems' LIMIT 1;
    SELECT user_id INTO it_dev_user_id FROM users WHERE username = 'it_dev' LIMIT 1;
    SELECT user_id INTO accounting_user_id FROM users WHERE username = 'accounting' LIMIT 1;
    SELECT user_id INTO payroll_user_id FROM users WHERE username = 'payroll' LIMIT 1;
    SELECT user_id INTO recruitment_user_id FROM users WHERE username = 'recruitment' LIMIT 1;
    SELECT user_id INTO benefits_user_id FROM users WHERE username = 'benefits' LIMIT 1;
    
    -- Insert user_roles
    INSERT INTO user_roles (user_role_id, user_id, role_id)
    VALUES
        (gen_random_uuid(), admin_user_id, admin_role_id),
        (gen_random_uuid(), user_user_id, employee_role_id),
        (gen_random_uuid(), manager1_user_id, manager_role_id),
        (gen_random_uuid(), manager2_user_id, manager_role_id),
        (gen_random_uuid(), finance_admin_user_id, finance_admin_role_id),
        (gen_random_uuid(), hr_admin_user_id, hr_admin_role_id),
        (gen_random_uuid(), employee1_user_id, employee_role_id),
        (gen_random_uuid(), employee2_user_id, employee_role_id),
        (gen_random_uuid(), employee3_user_id, employee_role_id),
        (gen_random_uuid(), employee4_user_id, employee_role_id),
        (gen_random_uuid(), readonly_user_id, readonly_role_id),
        (gen_random_uuid(), temp_admin_user_id, temp_admin_role_id),
        (gen_random_uuid(), sales_east_user_id, employee_role_id),
        (gen_random_uuid(), sales_west_user_id, employee_role_id),
        (gen_random_uuid(), it_systems_user_id, employee_role_id),
        (gen_random_uuid(), it_dev_user_id, employee_role_id),
        (gen_random_uuid(), accounting_user_id, employee_role_id),
        (gen_random_uuid(), payroll_user_id, employee_role_id),
        (gen_random_uuid(), recruitment_user_id, employee_role_id),
        (gen_random_uuid(), benefits_user_id, employee_role_id);
    
    -- Insert table_access for ADMIN role
-- Insert table_access for ADMIN role
INSERT INTO table_access (table_access_id, role_id, table_name, can_create, can_read, can_update, can_delete)
VALUES
    -- Core system tables
    (gen_random_uuid(), admin_role_id, 'users', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'roles', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'user_roles', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'table_access', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'restricted_columns', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'organizational_units', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'user_organizations', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'org_unit_column_access', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'cross_unit_permissions', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'permission_overrides', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'temporary_unit_access', true, true, true, true),
    
    -- Geography/location tables
    (gen_random_uuid(), admin_role_id, 'countries', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'regions', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'cities', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'streets', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'offices', true, true, true, true),
    
    -- User-related tables
    (gen_random_uuid(), admin_role_id, 'user_approval_status', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'titles', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'user_approval_comments', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'reports_to', true, true, true, true),
    
    -- Asset management tables
    (gen_random_uuid(), admin_role_id, 'asset_types', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'asset_conditions', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'valid_assets', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'assets', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'user_assets', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'tags', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'asset_tags', true, true, true, true),
    
    -- Product and related tables
    (gen_random_uuid(), admin_role_id, 'products', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'product_categories', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'brands', true, true, true, true),
    
    -- Status tables
    (gen_random_uuid(), admin_role_id, 'approval_status', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'fulfillment_status', true, true, true, true),
    
    -- Contact and personal information
    (gen_random_uuid(), admin_role_id, 'contact_info', true, true, true, true),
    (gen_random_uuid(), admin_role_id, 'homes', true, true, true, true),
    
    -- View access (for completeness)
    (gen_random_uuid(), admin_role_id, 'view_products', true, true, true, true);

    -- Insert table_access for EMPLOYEE role (basic access to users, products, assets)
INSERT INTO table_access (table_access_id, role_id, table_name, can_create, can_read, can_update, can_delete)
VALUES
    (gen_random_uuid(), employee_role_id, 'users', false, true, false, false),
    (gen_random_uuid(), employee_role_id, 'products', true, true, true, false),
    (gen_random_uuid(), employee_role_id, 'view_products', false, true, false, false),
    (gen_random_uuid(), employee_role_id, 'user_assets', true, true, true, false),
    (gen_random_uuid(), employee_role_id, 'assets', false, true, false, false),
    (gen_random_uuid(), employee_role_id, 'valid_assets', false, true, false, false),
    (gen_random_uuid(), employee_role_id, 'homes', true, true, true, false),
    (gen_random_uuid(), employee_role_id, 'contact_info', true, true, true, false),
    (gen_random_uuid(), employee_role_id, 'tags', false, true, false, false);

-- Insert table_access for MANAGER role (extended permissions)
INSERT INTO table_access (table_access_id, role_id, table_name, can_create, can_read, can_update, can_delete)
VALUES
    (gen_random_uuid(), manager_role_id, 'users', false, true, true, false),
    (gen_random_uuid(), manager_role_id, 'user_organizations', false, true, true, false),
    (gen_random_uuid(), manager_role_id, 'organizational_units', false, true, false, false),
    (gen_random_uuid(), manager_role_id, 'products', true, true, true, true),
    (gen_random_uuid(), manager_role_id, 'view_products', false, true, false, false),
    (gen_random_uuid(), manager_role_id, 'assets', true, true, true, false),
    (gen_random_uuid(), manager_role_id, 'valid_assets', false, true, false, false),
    (gen_random_uuid(), manager_role_id, 'user_assets', true, true, true, true),
    (gen_random_uuid(), manager_role_id, 'reports_to', true, true, true, true),
    (gen_random_uuid(), manager_role_id, 'homes', true, true, true, false),
    (gen_random_uuid(), manager_role_id, 'contact_info', true, true, true, false),
    (gen_random_uuid(), manager_role_id, 'tags', true, true, true, false),
    (gen_random_uuid(), manager_role_id, 'asset_tags', true, true, true, false);

-- Insert table_access for FINANCE_ADMIN role (financial-related tables)
INSERT INTO table_access (table_access_id, role_id, table_name, can_create, can_read, can_update, can_delete)
VALUES
    (gen_random_uuid(), finance_admin_role_id, 'users', false, true, false, false),
    (gen_random_uuid(), finance_admin_role_id, 'user_organizations', true, true, true, false),
    (gen_random_uuid(), finance_admin_role_id, 'organizational_units', false, true, false, false),
    (gen_random_uuid(), finance_admin_role_id, 'assets', true, true, true, true),
    (gen_random_uuid(), finance_admin_role_id, 'valid_assets', true, true, true, true),
    (gen_random_uuid(), finance_admin_role_id, 'user_assets', true, true, true, true),
    (gen_random_uuid(), finance_admin_role_id, 'products', true, true, true, true),
    (gen_random_uuid(), finance_admin_role_id, 'view_products', false, true, false, false),
    (gen_random_uuid(), finance_admin_role_id, 'asset_types', true, true, true, true),
    (gen_random_uuid(), finance_admin_role_id, 'asset_conditions', true, true, true, true),
    (gen_random_uuid(), finance_admin_role_id, 'product_categories', true, true, true, true),
    (gen_random_uuid(), finance_admin_role_id, 'brands', true, true, true, true),
    (gen_random_uuid(), finance_admin_role_id, 'tags', true, true, true, true),
    (gen_random_uuid(), finance_admin_role_id, 'asset_tags', true, true, true, true),
    (gen_random_uuid(), finance_admin_role_id, 'approval_status', true, true, true, true),
    (gen_random_uuid(), finance_admin_role_id, 'fulfillment_status', true, true, true, true);

-- Insert table_access for HR_ADMIN role (HR-related tables)
INSERT INTO table_access (table_access_id, role_id, table_name, can_create, can_read, can_update, can_delete)
VALUES
    (gen_random_uuid(), hr_admin_role_id, 'users', true, true, true, false),
    (gen_random_uuid(), hr_admin_role_id, 'user_organizations', true, true, true, false),
    (gen_random_uuid(), hr_admin_role_id, 'organizational_units', false, true, false, false),
    (gen_random_uuid(), hr_admin_role_id, 'reports_to', true, true, true, true),
    (gen_random_uuid(), hr_admin_role_id, 'titles', true, true, true, true),
    (gen_random_uuid(), hr_admin_role_id, 'offices', true, true, true, true),
    (gen_random_uuid(), hr_admin_role_id, 'user_approval_status', true, true, true, true),
    (gen_random_uuid(), hr_admin_role_id, 'user_approval_comments', true, true, true, true),
    (gen_random_uuid(), hr_admin_role_id, 'contact_info', true, true, true, true),
    (gen_random_uuid(), hr_admin_role_id, 'homes', true, true, true, true),
    (gen_random_uuid(), hr_admin_role_id, 'cross_unit_permissions', true, true, true, true),
    (gen_random_uuid(), hr_admin_role_id, 'temporary_unit_access', true, true, true, true);

-- Insert table_access for READONLY role (read-only access to most tables)
INSERT INTO table_access (table_access_id, role_id, table_name, can_create, can_read, can_update, can_delete)
VALUES
    (gen_random_uuid(), readonly_role_id, 'users', false, true, false, false),
    (gen_random_uuid(), readonly_role_id, 'roles', false, true, false, false),
    (gen_random_uuid(), readonly_role_id, 'organizational_units', false, true, false, false),
    (gen_random_uuid(), readonly_role_id, 'user_organizations', false, true, false, false),
    (gen_random_uuid(), readonly_role_id, 'products', false, true, false, false),
    (gen_random_uuid(), readonly_role_id, 'view_products', false, true, false, false),
    (gen_random_uuid(), readonly_role_id, 'assets', false, true, false, false),
    (gen_random_uuid(), readonly_role_id, 'valid_assets', false, true, false, false),
    (gen_random_uuid(), readonly_role_id, 'user_assets', false, true, false, false),
    (gen_random_uuid(), readonly_role_id, 'asset_types', false, true, false, false),
    (gen_random_uuid(), readonly_role_id, 'asset_conditions', false, true, false, false),
    (gen_random_uuid(), readonly_role_id, 'product_categories', false, true, false, false),
    (gen_random_uuid(), readonly_role_id, 'brands', false, true, false, false),
    (gen_random_uuid(), readonly_role_id, 'tags', false, true, false, false),
    (gen_random_uuid(), readonly_role_id, 'asset_tags', false, true, false, false),
    (gen_random_uuid(), readonly_role_id, 'reports_to', false, true, false, false);

-- Insert table_access for TEMP_ADMIN role (temporary admin access)
INSERT INTO table_access (table_access_id, role_id, table_name, can_create, can_read, can_update, can_delete)
VALUES
    (gen_random_uuid(), temp_admin_role_id, 'users', true, true, true, false),
    (gen_random_uuid(), temp_admin_role_id, 'roles', false, true, false, false),
    (gen_random_uuid(), temp_admin_role_id, 'user_roles', true, true, true, false),
    (gen_random_uuid(), temp_admin_role_id, 'organizational_units', true, true, true, false),
    (gen_random_uuid(), temp_admin_role_id, 'user_organizations', true, true, true, true),
    (gen_random_uuid(), temp_admin_role_id, 'products', true, true, true, true),
    (gen_random_uuid(), temp_admin_role_id, 'assets', true, true, true, true),
    (gen_random_uuid(), temp_admin_role_id, 'valid_assets', true, true, true, true),
    (gen_random_uuid(), temp_admin_role_id, 'user_assets', true, true, true, true),
    (gen_random_uuid(), temp_admin_role_id, 'temporary_unit_access', true, true, true, true),
    (gen_random_uuid(), temp_admin_role_id, 'cross_unit_permissions', true, true, true, true),
    (gen_random_uuid(), temp_admin_role_id, 'asset_types', true, true, true, false),
    (gen_random_uuid(), temp_admin_role_id, 'asset_conditions', true, true, true, false),
    (gen_random_uuid(), temp_admin_role_id, 'tags', true, true, true, false),
    (gen_random_uuid(), temp_admin_role_id, 'asset_tags', true, true, true, false);

    -- Insert organizational_units, starting with HQ
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'Company Headquarters', NULL, 0, 'Company_Headquarters', true, true, 'COMPANY', true)
    RETURNING organizational_unit_id INTO hq_id;

    -- Insert Finance Department and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'Finance Department', hq_id, 1, 'Company_Headquarters.Finance_Department', true, true, 'DEPARTMENT', true)
    RETURNING organizational_unit_id INTO finance_dept_id;

    -- Insert HR Department and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'HR Department', hq_id, 1, 'Company_Headquarters.HR_Department', true, true, 'DEPARTMENT', true)
    RETURNING organizational_unit_id INTO hr_dept_id;

    -- Insert Sales Department and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'Sales Department', hq_id, 1, 'Company_Headquarters.Sales_Department', true, true, 'DEPARTMENT', true)
    RETURNING organizational_unit_id INTO sales_dept_id;

    -- Insert IT Department and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'IT Department', hq_id, 1, 'Company_Headquarters.IT_Department', true, true, 'DEPARTMENT', true)
    RETURNING organizational_unit_id INTO it_dept_id;
    
    -- Insert Accounting Team and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'Accounting Team', finance_dept_id, 2, 'Company_Headquarters.Finance_Department.Accounting_Team', true, true, 'TEAM', true)
    RETURNING organizational_unit_id INTO accounting_team_id;
    
    -- Insert Payroll Team and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'Payroll Team', finance_dept_id, 2, 'Company_Headquarters.Finance_Department.Payroll_Team', true, true, 'TEAM', true)
    RETURNING organizational_unit_id INTO payroll_team_id;
    
    -- Insert Recruitment Team and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'Recruitment Team', hr_dept_id, 2, 'Company_Headquarters.HR_Department.Recruitment_Team', true, true, 'TEAM', true)
    RETURNING organizational_unit_id INTO recruitment_team_id;
    
    -- Insert Benefits Team and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'Benefits Team', hr_dept_id, 2, 'Company_Headquarters.HR_Department.Benefits_Team', true, true, 'TEAM', true)
    RETURNING organizational_unit_id INTO benefits_team_id;
    
    -- Insert East Region and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'East Region', sales_dept_id, 2, 'Company_Headquarters.Sales_Department.East_Region', true, true, 'REGION', true)
    RETURNING organizational_unit_id INTO east_region_id;
    
    -- Insert West Region and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'West Region', sales_dept_id, 2, 'Company_Headquarters.Sales_Department.West_Region', true, true, 'REGION', true)
    RETURNING organizational_unit_id INTO west_region_id;
    
    -- Insert Systems Team and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'Systems Team', it_dept_id, 2, 'Company_Headquarters.IT_Department.Systems_Team', true, true, 'TEAM', true)
    RETURNING organizational_unit_id INTO systems_team_id;
    
    -- Insert Development Team and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'Development Team', it_dept_id, 2, 'Company_Headquarters.IT_Department.Development_Team', true, true, 'TEAM', true)
    RETURNING organizational_unit_id INTO dev_team_id;
    
    -- Insert Northeast Branch and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'Northeast Branch', east_region_id, 3, 'Company_Headquarters.Sales_Department.East_Region.Northeast_Branch', true, true, 'BRANCH', true)
    RETURNING organizational_unit_id INTO northeast_branch_id;
    
    -- Insert Southeast Branch and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'Southeast Branch', east_region_id, 3, 'Company_Headquarters.Sales_Department.East_Region.Southeast_Branch', true, true, 'BRANCH', true)
    RETURNING organizational_unit_id INTO southeast_branch_id;
    
    -- Insert Northwest Branch and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'Northwest Branch', west_region_id, 3, 'Company_Headquarters.Sales_Department.West_Region.Northwest_Branch', true, true, 'BRANCH', true)
    RETURNING organizational_unit_id INTO northwest_branch_id;
    
    -- Insert Southwest Branch and get its ID
    INSERT INTO organizational_units (organizational_unit_id, name, parent_id, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active)
    VALUES 
        (gen_random_uuid(), 'Southwest Branch', west_region_id, 3, 'Company_Headquarters.Sales_Department.West_Region.Southwest_Branch', true, true, 'BRANCH', true)
    RETURNING organizational_unit_id INTO southwest_branch_id;
    
    -- Insert user_organizations - assign users to organizational units with appropriate roles
    INSERT INTO user_organizations (user_organization_id, user_id, organizational_unit_id, role_id, is_unit_manager, start_date, created_by)
    VALUES
        -- Admin at HQ
        (gen_random_uuid(), admin_user_id, hq_id, admin_role_id, true, CURRENT_TIMESTAMP, admin_user_id),
        -- Regular user at HQ
        (gen_random_uuid(), user_user_id, hq_id, employee_role_id, false, CURRENT_TIMESTAMP, admin_user_id),
        -- Department managers
        (gen_random_uuid(), manager1_user_id, finance_dept_id, manager_role_id, true, CURRENT_TIMESTAMP, admin_user_id),
        (gen_random_uuid(), manager2_user_id, hr_dept_id, manager_role_id, true, CURRENT_TIMESTAMP, admin_user_id),
        -- Department admins
        (gen_random_uuid(), finance_admin_user_id, finance_dept_id, finance_admin_role_id, false, CURRENT_TIMESTAMP, admin_user_id),
        (gen_random_uuid(), hr_admin_user_id, hr_dept_id, hr_admin_role_id, false, CURRENT_TIMESTAMP, admin_user_id),
        -- Employees in various teams
        (gen_random_uuid(), employee1_user_id, accounting_team_id, employee_role_id, false, CURRENT_TIMESTAMP, manager1_user_id),
        (gen_random_uuid(), employee2_user_id, payroll_team_id, employee_role_id, false, CURRENT_TIMESTAMP, manager1_user_id),
        (gen_random_uuid(), employee3_user_id, recruitment_team_id, employee_role_id, false, CURRENT_TIMESTAMP, manager2_user_id),
        (gen_random_uuid(), employee4_user_id, benefits_team_id, employee_role_id, false, CURRENT_TIMESTAMP, manager2_user_id),
        -- Specialized team members
        (gen_random_uuid(), accounting_user_id, accounting_team_id, employee_role_id, true, CURRENT_TIMESTAMP, manager1_user_id),
        (gen_random_uuid(), payroll_user_id, payroll_team_id, employee_role_id, true, CURRENT_TIMESTAMP, manager1_user_id),
        (gen_random_uuid(), recruitment_user_id, recruitment_team_id, employee_role_id, true, CURRENT_TIMESTAMP, manager2_user_id),
        (gen_random_uuid(), benefits_user_id, benefits_team_id, employee_role_id, true, CURRENT_TIMESTAMP, manager2_user_id),
        -- Sales team members
        (gen_random_uuid(), sales_east_user_id, east_region_id, employee_role_id, true, CURRENT_TIMESTAMP, admin_user_id),
        (gen_random_uuid(), sales_west_user_id, west_region_id, employee_role_id, true, CURRENT_TIMESTAMP, admin_user_id),
        -- IT team members
        (gen_random_uuid(), it_systems_user_id, systems_team_id, employee_role_id, true, CURRENT_TIMESTAMP, admin_user_id),
        (gen_random_uuid(), it_dev_user_id, dev_team_id, employee_role_id, true, CURRENT_TIMESTAMP, admin_user_id),
        -- Readonly user at company level
        (gen_random_uuid(), readonly_user_id, hq_id, readonly_role_id, false, CURRENT_TIMESTAMP, admin_user_id),
        -- Temp admin user with elevated access
        (gen_random_uuid(), temp_admin_user_id, hq_id, temp_admin_role_id, false, CURRENT_TIMESTAMP, admin_user_id);
END $$;