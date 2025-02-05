CREATE DATABASE CP_Server_Administrator_WA;
USE CP_Server_Administrator_WA;

CREATE TABLE wa_users (
  wa_users_id INT PRIMARY KEY AUTO_INCREMENT,
  wa_users_username VARCHAR(60),
  wa_users_password VARCHAR(130),
  wa_users_role VARCHAR(60)
);

-- INSERT INTO wa_users (wa_users_username, wa_users_password, wa_users_role) VALUES ("trilx123","9a835b7eece9ea09bfc80b63d15b94aee929eac524544813da1962bc35081fbaea7698c84b73b7b3d7c65ead23d7abbf0d8e25e183e50f6a1f1e96f97d712afd", "admin");
-- INSERT INTO wa_users (wa_users_username, wa_users_password, wa_users_role) VALUES ("long123","556c45b340635b61ab3a99a282d5c339115fe9e636d859edc5cdc9dabcbb701198f50c5e204dc0e3393f7c54b6116525d12e4d84690081761b42632c87002f2c", "admin");
-- SELECT * FROM wa_users;

CREATE TABLE  ssh_keys (
    sk_key_id INT PRIMARY KEY AUTO_INCREMENT,
    sk_key_name varchar(60),
    sk_private_key text,
    creator_id INT,
    FOREIGN KEY (creator_id) references wa_users(wa_users_id)
);

CREATE TABLE ssh_connections (
    sc_connection_id INT PRIMARY KEY AUTO_INCREMENT,
    sc_username VARCHAR(60),
    sc_password varchar(50),
    sc_host varchar(60),
    sc_hostname varchar(60),
    sc_port INT,
    creator_id INT,
    ssh_key_id INT,
    FOREIGN KEY (creator_id) references wa_users(wa_users_id),
    FOREIGN KEY (ssh_key_id) references ssh_keys(sk_key_id)
);

CREATE TABLE package_installed (
  pkg_id INT PRIMARY KEY AUTO_INCREMENT,
  pkg_name VARCHAR(60),
  pkg_date DATETIME,
  pkg_host_id INT,
  FOREIGN KEY (pkg_host_id) references ssh_connections(sc_connection_id)
  ON DELETE CASCADE
);

CREATE TABLE event_web (
  ev_web_id INT PRIMARY KEY AUTO_INCREMENT,
  ev_web_type VARCHAR(60),
  ev_web_description VARCHAR(300),
  ev_web_timestamp DATETIME,
  ev_web_creator_id INT,
  FOREIGN KEY (ev_web_creator_id) references wa_users(wa_users_id) ON DELETE CASCADE
  
);

CREATE TABLE  invent_group (
    invent_group_id INT PRIMARY KEY AUTO_INCREMENT,
    invent_group_name varchar(60),
);

ALTER TABLE ssh_connections ADD group_id INT;
FOREIGN KEY (group_id) references invent_group(invent_group_id);


CREATE TABLE snmp_credential (
    snmp_id INT PRIMARY KEY AUTO_INCREMENT,
    snmp_auth_username varchar(60),
    snmp_auth_password varchar(60),
    snmp_priv_password varchar(60),
    snmp_connection_id INT,
    FOREIGN KEY (snmp_connection_id) references ssh_connections(sc_connection_id)
    ON DELETE CASCADE
);

CREATE TABLE ssh_connections_information (
    sc_info_id INT PRIMARY KEY AUTO_INCREMENT,
    sc_info_osname varchar(60),
    sc_info_osversion varchar(60),
    sc_info_installdate DATETIME,
    sc_info_serial varchar(60),
    sc_info_hostname varchar(60),
    sc_info_connection_id int,
    FOREIGN KEY (sc_info_connection_id) references ssh_connections(sc_connection_id)
    ON DELETE CASCADE
);


CREATE TABLE ssh_connection_alert (
    sca_id INT PRIMARY KEY,
    sca_connection_name varchar(60),
    sca_alert_pri varchar(60),
    FOREIGN KEY (sca_id) references ssh_connections(sc_connection_id)
    ON DELETE CASCADE
);

CREATE TABLE `templates` (
  `template_id` int NOT NULL AUTO_INCREMENT,
  `template_name` varchar(60) DEFAULT NULL,
  `template_description` varchar(200) DEFAULT NULL,
  `ssh_key_id` int NULL,
  `filepath` varchar(100) DEFAULT NULL,
  `arguments` text,
  `alert` tinyint(1) NOT NULL,
  `user_id` int DEFAULT NULL,
  PRIMARY KEY (`template_id`),
  KEY `ssh_key_id` (`ssh_key_id`),
  CONSTRAINT `templates_ibfk_1` FOREIGN KEY (`ssh_key_id`) REFERENCES `ssh_keys` (`sk_key_id`) ON DELETE CASCADE
)