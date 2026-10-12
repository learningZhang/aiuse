# 基金涨跌邮件推送 部署说明

## 文件清单

| 文件 | 说明 |
|------|------|
| fund_email_push.py | 主推送脚本 |
| my_funds.json | 基金持仓数据 |
| email_config.json | 邮箱SMTP配置 |

## 部署步骤

### 1. 安装依赖
```bash
pip3 install requests
```

### 2. 修改配置
- **email_config.json**：填写你的邮箱SMTP信息
- **my_funds.json**：增减基金、修改持有金额

### 3. 测试运行
```bash
python3 fund_email_push.py
```
看到 `✅ 邮件发送成功！` 即可

### 4. 设置定时任务（Linux/macOS）
```bash
crontab -e
```
添加以下三行（工作日 11:45、14:45、15:45 执行）：
```
45 11 * * 1-5 python3 /你的路径/fund_email_push.py >> /你的路径/push.log 2>&1
45 14 * * 1-5 python3 /你的路径/fund_email_push.py >> /你的路径/push.log 2>&1
45 15 * * 1-5 python3 /你的路径/fund_email_push.py >> /你的路径/push.log 2>&1
```

### 5. Windows 任务计划程序
1. 打开"任务计划程序"
2. 创建基本任务 → 触发器选"每天"
3. 操作选"启动程序"，程序填 `python3`，参数填脚本完整路径
4. 分别创建 11:45、14:45、15:45 三个任务

## 修改持仓金额
编辑 my_funds.json，修改 amount 字段即可：
```json
{"name": "基金名称", "code": "代码", "amount": 持有金额}
```

## 计算逻辑
估算今日收益 = 持有金额 × 估算涨跌幅
