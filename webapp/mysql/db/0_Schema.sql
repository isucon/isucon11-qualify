DROP TABLE IF EXISTS `isu`;
CREATE TABLE `isu` (
  `id` VARCHAR(255) PRIMARY KEY,
  `name` VARCHAR(255) NOT NULL,
  `image` LONGBLOB,
  `catalog_id` VARCHAR(255) NOT NULL,
  `user_id` BIGINT NOT NULL,
  `is_deleted` TINYINT(1) DEFAULT FALSE,
  `created_at` DATETIME(6) NOT NULL,
  `updated_at` DATETIME(6) NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;

-- condition は結局 JSON 何だっけ？
DROP TABLE IF EXISTS `isu_log`;
CREATE TABLE `isu_log` (
  `isu_id` VARCHAR(255),
  `timestamp` DATETIME(6),
  `condition` JSON NOT NULL,
-- `is_dirty` TINYINT(1) DEFAULT FALSE,
-- `is_overweight` TINYINT(1) DEFAULT FALSE,
-- `is_broken` TINYINT(1) DEFAULT FALSE,
  `message` VARCHAR(255) NOT NULL,
  `created_at` DATETIME(6) NOT NULL,
  PRIMARY KEY(`isu_id`,`timestamp`)
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;

DROP TABLE IF EXISTS `graph`;
CREATE TABLE `graph` (
  `isu_id` VARCHAR(255),
  `start_at` DATETIME(6),
  `data` JSON NOT NULL,
  `created_at` DATETIME(6) NOT NULL,
  `updated_at` DATETIME(6) NOT NULL,
  PRIMARY KEY(`isu_id`,`start_at`)
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;

DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `name` VARCHAR(255) NOT NULL UNIQUE,
  `created_at` DATETIME(6) NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;

-- 以下の図より、平常時とベンチ時とで何かを切り替えるのはベンチ側であり App 側は何もしなくて良さそうなので、以下のテーブルは要らないかも？
-- https://docs.google.com/presentation/d/1yzeYYi39hW66D42u7TrY7zLCz0N1RQyatBv8n49b5Xw/edit#slide=id.gdd6b21deeb_0_5
DROP TABLE IF EXISTS `isu_association_config`;
CREATE TABLE `isu_association_config` (
  `name` VARCHAR(255) PRIMARY KEY,
  `url` VARCHAR(255) NOT NULL UNIQUE
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;
