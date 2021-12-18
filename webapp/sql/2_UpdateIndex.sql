ALTER TABLE `isu` ADD INDEX `idx_1` (`jia_user_id`, `jia_isu_uuid`);
ALTER TABLE `isu_condition` ADD INDEX `idx_1` (`jia_isu_uuid`);
ALTER TABLE `isu_condition` ADD INDEX `idx_2` (`jia_isu_uuid`, `timestamp`);
ALTER TABLE `isu_association_config` ADD INDEX `idx_1` (`name`);
