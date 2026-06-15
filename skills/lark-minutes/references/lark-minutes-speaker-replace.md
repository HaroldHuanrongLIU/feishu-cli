# minutes +speaker-replace

> **前置条件：** 先阅读 [`../lark-shared/SKILL.md`](../../lark-shared/SKILL.md) 了解认证、全局参数和安全规则。

替换妙记逐字稿中的说话人身份：把妙记逐字稿里"原说话人"对应的所有发言段，重新归属到"新说话人"。常用于解决妙记自动识别错说话人，或需要把外部/非飞书说话人改绑到正确飞书用户的场景。

本 skill 对应 shortcut：`lark-cli minutes +speaker-replace`。

## 典型触发表达

- "把这条妙记里 A 的发言改成 B"
- "妙记说话人识别错了，帮我把张三的部分换成李四"
- "把妙记里外部说话人 / 非飞书说话人的发言改成某个飞书用户"
- "妙记说话人修改 / 替换 / 重新归属"

## 工作流（重要）

对外参数始终是 **`--from-speaker-id`**。用户若已知不透明 `speaker_id`，直接传入；若只知道展示名（如「说话人1」），也传给 `--from-speaker-id`，**CLI 会在内部请求说话人列表 HTTP 接口**，按展示名解析出真正的 `speaker_id` 后再调用替换 API（GetSpeakers 不对外暴露为独立命令）。

标准流程：

1. 用户提供原说话人展示名或 `speaker_id`，以及新说话人姓名。
2. 调用 `+speaker-replace`，`--from-speaker-id` 传展示名或已知 id。
3. CLI 内部 GET speakerlist → 若入参已是 `speaker_id` 则直接使用，否则按展示名匹配。
4. 若有多个同名说话人，**不要擅自挑选**：CLI 会报错并列出匹配的 `speaker_id`；提醒用户在妙记里查看每个同名说话人各自说过的内容，由用户确认后传入精确的 `--from-speaker-id`。
5. 新说话人解析成 `ou_` 开头的 open_id（姓名先用 [lark-contact](../../lark-contact/SKILL.md) 解析）。

## 命令示例

```bash
# 用户只知道展示名
lark-cli minutes +speaker-replace \
  --minute-token obcnxxxxxxxxxxxxxxxxxxxx \
  --from-speaker-id "说话人1" \
  --to-user-id ou_new_speaker_open_id

# 用户已知不透明 speaker_id
lark-cli minutes +speaker-replace \
  --minute-token obcnxxxxxxxxxxxxxxxxxxxx \
  --from-speaker-id ENCRYPTED_TOKEN_ABC \
  --to-user-id ou_new_speaker_open_id
```

## 参数

| 参数 | 必填 | 说明 |
|------|------|------|
| `--minute-token <token>` | 是 | 妙记的唯一标识，可从妙记 URL 末尾路径提取 |
| `--from-speaker-id <id>` | 是 | 被替换的原说话人：不透明 `speaker_id`，或展示名（CLI 内部解析，支持外部/非飞书说话人） |
| `--to-user-id <ou_xxx>` | 是 | 新的说话人，**必须是 `ou_` 开头的 open_id**，不支持用户名 |

> **重要**：
> - 始终使用 `--from-speaker-id`；展示名和真实 id 都通过这个参数传入，由 CLI 内部区分。
> - `--to-user-id` 仅支持 `ou_` 开头的 open_id，**不支持直接传姓名**；如果用户只给了姓名，请先用 [lark-contact](../../lark-contact/SKILL.md) 把姓名解析成 `open_id`。
> - 存在一个隐藏的历史参数 `--from-user-id`（飞书说话人的 open_id），仅为向后兼容保留。

## 认证与权限

- 所需 scope：`minutes:minutes:readonly`（内部解析说话人）、`minutes:minutes:update`（执行替换）。

## 输出结果

| 字段 | 说明 |
|------|------|
| `minute_token` | 被修改的妙记 Token，与输入的 `--minute-token` 一致 |
| `from_speaker_id` | 实际用于替换的不透明说话人标识 |
| `from_speaker_input` | 仅当入参为展示名且被解析时，回显原始 `--from-speaker-id` 入参 |
| `to_user_id` | 替换后的新说话人 open_id，与输入的 `--to-user-id` 一致 |

## 参考

- [lark-minutes](../SKILL.md) -- 妙记相关功能说明
- [lark-shared](../../lark-shared/SKILL.md) -- 认证和全局参数
