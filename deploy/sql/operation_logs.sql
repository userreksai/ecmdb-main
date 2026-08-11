CREATE TABLE IF NOT EXISTS `operation_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键',
  `account` VARCHAR(128) NOT NULL COMMENT '操作账号',
  `operation_model` VARCHAR(128) NOT NULL COMMENT '操作模型唯一标识',
  `operation_type` VARCHAR(16) NOT NULL COMMENT '操作类型：CREATE/UPDATE/DELETE/IMPORT',
  `original_data` JSON NULL COMMENT '原数据',
  `modified_data` JSON NULL COMMENT '修改后的数据',
  `operation_time` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '操作日期',
  `modified_count` BIGINT NOT NULL DEFAULT 0 COMMENT '修改条数',
  PRIMARY KEY (`id`),
  KEY `idx_operation_log_account` (`account`),
  KEY `idx_operation_log_model` (`operation_model`),
  KEY `idx_operation_log_type` (`operation_type`),
  KEY `idx_operation_log_time` (`operation_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='CMDB 模型数据操作日志（默认保留一个月）';
