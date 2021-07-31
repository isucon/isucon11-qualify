ALTER TABLE isu ADD CONSTRAINT `isu_user_id` FOREIGN KEY (jia_user_id) REFERENCES user(jia_user_id);
-- ALTER TABLE isu_condition ADD CONSTRAINT `isu_condition_isu_uuid` FOREIGN KEY (jia_isu_uuid) REFERENCES isu(jia_isu_uuid);
