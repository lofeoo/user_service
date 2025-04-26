-- +goose Up
-- +goose StatementBegin
CREATE TABLE `users` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` varchar(24) NOT NULL COMMENT '用户名',
  `password` varchar(128) NOT NULL COMMENT '密码',
  `email` varchar(64) NOT NULL COMMENT '邮箱',
  `phone` varchar(20) NOT NULL DEFAULT '' COMMENT '手机号',
  `avatar` varchar(255) NOT NULL DEFAULT '' COMMENT '头像URL',
  `real_name` varchar(32) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `gender` varchar(10) NOT NULL DEFAULT '' COMMENT '性别',
  `birthday` varchar(10) NOT NULL DEFAULT '' COMMENT '生日',
  `address` varchar(255) NOT NULL DEFAULT '' COMMENT '地址',
  `bio` text COMMENT '个人简介',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态(0:禁用,1:启用)',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`),
  UNIQUE KEY `idx_email` (`email`),
  KEY `idx_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

CREATE TABLE `oauth_accounts` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint(20) UNSIGNED NOT NULL COMMENT '用户ID',
  `provider` varchar(32) NOT NULL COMMENT '提供商(github,google等)',
  `provider_user_id` varchar(128) NOT NULL COMMENT '提供商用户ID',
  `access_token` varchar(255) NOT NULL DEFAULT '' COMMENT '访问令牌',
  `refresh_token` varchar(255) NOT NULL DEFAULT '' COMMENT '刷新令牌',
  `expires_at` timestamp NULL DEFAULT NULL COMMENT '过期时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_provider_uid` (`provider`, `provider_user_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='OAuth账号表';

CREATE TABLE `tokens` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint(20) UNSIGNED NOT NULL COMMENT '用户ID',
  `refresh_token` varchar(255) NOT NULL COMMENT '刷新令牌',
  `refresh_token_expires_at` timestamp NOT NULL COMMENT '刷新令牌过期时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_refresh_token` (`refresh_token`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='令牌表';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `tokens`;
DROP TABLE IF EXISTS `oauth_accounts`;
DROP TABLE IF EXISTS `users`;
-- +goose StatementEnd 