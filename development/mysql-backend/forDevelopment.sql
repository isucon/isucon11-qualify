ALTER TABLE isu ADD CONSTRAINT `isu_user_id` FOREIGN KEY (user_id) REFERENCES user(id);
ALTER TABLE isu_log ADD CONSTRAINT `isu_log_isu_uuid` FOREIGN KEY (isu_id) REFERENCES isu(uuid);
ALTER TABLE graph ADD CONSTRAINT `graph_isu_uuid` FOREIGN KEY (isu_id) REFERENCES isu(uuid);
