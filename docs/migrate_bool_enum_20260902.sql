-- 布尔语义字段迁移脚本（在 migrate_status_20260901.sql 基础上执行）
-- 目标：以下 0否/1是 字段不再使用 0，统一改为 1否/2是
-- 映射规则（整体 +1）：
--   activity.auto_audit:    旧 0否(需审核) 1是(免审核)  -> 新 1否(需审核) 2是(免审核)
--   activity.is_official:   旧 0否(用户发布) 1是(平台官方) -> 新 1否(用户发布) 2是(平台官方)
--   activity.warn_notified: 旧 0否(未通知) 1是(已通知)   -> 新 1否(未通知) 2是(已通知)
--   message.is_read:        旧 0否(未读) 1是(已读)       -> 新 1否(未读) 2是(已读)
--   user.merchant_flag:     旧 0无(未开通) 1有(已开通)    -> 新 1无(未开通) 2有(已开通)

START TRANSACTION;

-- 活动布尔字段
UPDATE activity SET auto_audit = auto_audit + 1,
                    is_official = is_official + 1,
                    warn_notified = warn_notified + 1;

-- 消息已读
UPDATE message SET is_read = is_read + 1;

-- 用户入驻权限
UPDATE user SET merchant_flag = merchant_flag + 1;

COMMIT;

-- 验证：应无任何相关字段为 0 的记录
SELECT 'activity.auto_audit 校验' AS t, COUNT(*) AS zero_cnt FROM activity WHERE auto_audit = 0
UNION ALL SELECT 'activity.is_official 校验', COUNT(*) FROM activity WHERE is_official = 0
UNION ALL SELECT 'activity.warn_notified 校验', COUNT(*) FROM activity WHERE warn_notified = 0
UNION ALL SELECT 'message.is_read 校验', COUNT(*) FROM message WHERE is_read = 0
UNION ALL SELECT 'user.merchant_flag 校验', COUNT(*) FROM user WHERE merchant_flag = 0;
