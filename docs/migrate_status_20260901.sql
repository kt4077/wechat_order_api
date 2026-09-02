-- 状态值重构迁移脚本
-- 目标：所有状态字段不再使用 0，改为从 1 开始
-- 映射规则（整体 +1）：
--   signup.status:        旧 0待审核 1通过 2驳回 3撤销 -> 新 1待审核 2通过 3驳回 4撤销
--   merchant_apply.status: 旧 0待审核 1通过 2驳回        -> 新 1待审核 2通过 3驳回
--   activity.status:      旧 0草稿 1未开始 2报名中 3结束 4下架 5待审核 6驳回
--                          -> 新 1草稿 2未开始 3报名中 4结束 5下架 6待审核 7驳回
--   user.role:            旧 0普通 1入驻 2超管            -> 新 1普通 2入驻 3超管
--   user.gender:          旧 0未知 1男 2女                -> 新 1未知 2男 3女
--   message.type:         旧 0系统 1报名 2审核 3入驻 4活动 -> 新 1系统 2报名 3审核 4入驻 5活动

START TRANSACTION;

-- 报名状态
UPDATE signup SET status = status + 1;

-- 入驻申请状态
UPDATE merchant_apply SET status = status + 1;

-- 活动状态
UPDATE activity SET status = status + 1;

-- 用户角色与性别
UPDATE user SET role = role + 1, gender = gender + 1;

-- 消息类型
UPDATE message SET type = type + 1;

COMMIT;

-- 验证：应无任何 status/role/type 为 0 的记录
SELECT 'signup 状态校验' AS t, COUNT(*) AS zero_cnt FROM signup WHERE status = 0
UNION ALL SELECT 'merchant_apply 状态校验', COUNT(*) FROM merchant_apply WHERE status = 0
UNION ALL SELECT 'activity 状态校验', COUNT(*) FROM activity WHERE status = 0
UNION ALL SELECT 'user 角色校验', COUNT(*) FROM user WHERE role = 0
UNION ALL SELECT 'user 性别校验', COUNT(*) FROM user WHERE gender = 0
UNION ALL SELECT 'message 类型校验', COUNT(*) FROM message WHERE type = 0;
