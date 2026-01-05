-- 用户表结构定义
-- 符合MySQL最佳实践的建表语句
-- 请确保在正确的数据库中执行此脚本

-- 创建用户表
CREATE TABLE IF NOT EXISTS `users` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID，主键',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` timestamp NULL DEFAULT NULL COMMENT '软删除时间戳',
    `name` varchar(100) NOT NULL COMMENT '用户姓名',
    `email` varchar(100) NOT NULL COMMENT '用户邮箱',
    `age` int DEFAULT NULL COMMENT '用户年龄',
    `status` tinyint NOT NULL DEFAULT 1 COMMENT '用户状态 1-正常 0-禁用',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_users_email` (`email`),
    KEY `idx_users_deleted_at` (`deleted_at`),
    KEY `idx_users_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 创建索引优化查询性能
-- 为邮箱字段创建唯一索引，确保邮箱唯一性
ALTER TABLE `users` ADD CONSTRAINT `uk_users_email` UNIQUE (`email`) COMMENT '邮箱唯一约束';

-- 添加示例数据（可选）
-- INSERT INTO `users` (`name`, `email`, `age`, `status`) VALUES 
-- ('张三', 'zhangsan@example.com', 25, 1),
-- ('李四', 'lisi@example.com', 30, 1),
-- ('王五', 'wangwu@example.com', 28, 1);