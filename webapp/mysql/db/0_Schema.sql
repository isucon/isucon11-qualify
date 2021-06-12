DROP TABLE IF EXISTS `isu`;
CREATE TABLE `isu` (
  `uuid` CHAR(36) PRIMARY KEY,
  `name` VARCHAR(255) NOT NULL,
  `image` LONGBLOB,
  `catalog_id` VARCHAR(255) NOT NULL,
  `user_id` BIGINT NOT NULL,
  `is_deleted` TINYINT(1) NOT NULL DEFAULT FALSE,
  `created_at` DATETIME(6) NOT NULL,
  `updated_at` DATETIME(6) NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;

DROP TABLE IF EXISTS `isu_log`;
CREATE TABLE `isu_log` (
  `isu_id` VARCHAR(255),
  `timestamp` DATETIME,
  `condition` VARCHAR(255) NOT NULL,
  `message` VARCHAR(255) NOT NULL,
  `created_at` DATETIME(6) NOT NULL,
  PRIMARY KEY(`isu_id`,`timestamp`)
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;

DROP TABLE IF EXISTS `graph`;
CREATE TABLE `graph` (
  `isu_id` VARCHAR(255),
  `start_at` DATETIME,
  `data` JSON NOT NULL,
  `created_at` DATETIME(6) NOT NULL,
  `updated_at` DATETIME(6) NOT NULL,
  PRIMARY KEY(`isu_id`,`start_at`)
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;

DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `name` VARCHAR(255) NOT NULL UNIQUE,
  `created_at` DATETIME(6) NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;

DROP TABLE IF EXISTS `isu_association_config`;
CREATE TABLE `isu_association_config` (
  `name` VARCHAR(255) PRIMARY KEY,
  `url` VARCHAR(255) NOT NULL UNIQUE
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;
