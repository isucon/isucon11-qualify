ALTER TABLE isu ADD CONSTRAINT `isu_user_id` FOREIGN KEY (jia_user_id) REFERENCES user(jia_user_id);
ALTER TABLE isu_log ADD CONSTRAINT `isu_log_isu_uuid` FOREIGN KEY (jia_isu_uuid) REFERENCES isu(jia_isu_uuid);
ALTER TABLE graph ADD CONSTRAINT `graph_isu_uuid` FOREIGN KEY (jia_isu_uuid) REFERENCES isu(jia_isu_uuid);
