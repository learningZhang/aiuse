#!/usr/bin/env python3
"""基金涨跌邮件推送脚本 - 精简版（基金名称、持有金额、涨跌幅、估算收益）"""

import json
import smtplib
import requests
import logging
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from datetime import datetime

# 日志配置
logging.basicConfig(
    filename="./fund_push.log",
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(message)s"
)

# 读取基金列表（含持仓数据）
with open("./my_funds.json", "r") as f:
    funds = json.load(f)

# 读取邮件配置
with open("./email_config.json", "r") as f:
    email_cfg = json.load(f)

headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    'Referer': 'https://h5.1234567.com.cn/'
}

# 批量获取基金估值数据（新接口）
def fetch_fund_valuations(fund_list):
    """通过 FundValuationLast 接口批量获取基金估值"""
    codes = ",".join(f["code"] for f in fund_list)
    urls = [
        "https://fundcomapi.tiantianfunds.com/mm/newCore/FundValuationLast",
        "https://fundcomapi.eastmoney.com/mm/newCore/FundValuationLast",
    ]
    params = {
        "FCODES": codes,
        "FIELDS": "FCODE,SHORTNAME,GSZZL,GZTIME,GSZ,NAV,PDATE"
    }
    for url in urls:
        try:
            resp = requests.get(url, params=params, headers=headers, timeout=15)
            resp.raise_for_status()
            data = resp.json()
            if data.get("success") and data.get("errorCode") == 0:
                return {item["FCODE"]: item for item in data.get("data", [])}
        except Exception as e:
            logging.warning(f"请求 {url} 失败: {e}")
            continue
    return {}

valuation_map = fetch_fund_valuations(funds)

# 组装结果
results = []
for fund in funds:
    name = fund["name"]
    code = fund["code"]
    amount = fund.get("amount", 0)
    val = valuation_map.get(code, {})
    gszzl_raw = val.get("GSZZL")
    gztime = val.get("GZTIME", "-")

    if gszzl_raw is not None:
        gszzl = float(gszzl_raw)
        est_profit = amount * gszzl / 100
        results.append({
            "name": name, "code": code, "amount": amount,
            "gszzl": gszzl,
            "gszzl_str": f"{gszzl:+.2f}%",
            "est_profit": est_profit,
            "est_profit_str": f"{est_profit:+,.2f}",
            "gztime": gztime,
        })
    else:
        results.append({
            "name": name, "code": code, "amount": amount,
            "gszzl": 0, "gszzl_str": "暂无估值",
            "est_profit": 0, "est_profit_str": "-",
            "gztime": "-",
        })

# 按涨跌幅排序（从高到低）
results.sort(key=lambda x: x["gszzl"], reverse=True)

# 统计
total_amount = sum(r["amount"] for r in results)
total_est_profit = sum(r["est_profit"] for r in results)

up_count = sum(1 for r in results if r["gszzl"] > 0)
down_count = sum(1 for r in results if r["gszzl"] < 0)
flat_count = sum(1 for r in results if r["gszzl"] == 0)

today = datetime.now().strftime("%Y-%m-%d")
gztime = results[0]["gztime"] if results else today

# 找领涨领跌
top = results[0] if results else None
bottom = results[-1] if results else None

# 按收益排序找收益贡献最大/最小
results_by_profit = sorted(results, key=lambda x: x["est_profit"], reverse=True)
top_profit = results_by_profit[0] if results_by_profit else None
bottom_profit = results_by_profit[-1] if results_by_profit else None

# ===== 构建HTML邮件 =====
est_profit_color = "#e74c3c" if total_est_profit >= 0 else "#27ae60"

html = f"""
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>
  body {{ font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; margin: 0; padding: 16px; }}
  .container {{ max-width: 700px; margin: 0 auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 12px rgba(0,0,0,0.08); }}
  .header {{ background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: #fff; padding: 24px 24px 16px; }}
  .header h1 {{ margin: 0 0 4px 0; font-size: 20px; }}
  .header .sub {{ font-size: 12px; opacity: 0.85; }}
  .overview {{ display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; padding: 16px 20px; border-bottom: 1px solid #f0f0f0; }}
  .ov-item {{ text-align: center; padding: 8px; border-radius: 8px; background: #fafafa; }}
  .ov-item .num {{ font-size: 22px; font-weight: 700; }}
  .ov-item .label {{ font-size: 11px; color: #999; margin-top: 2px; }}
  .up {{ color: #e74c3c; }}
  .down {{ color: #27ae60; }}
  .flat {{ color: #999; }}
  .count-bar {{ display: flex; justify-content: center; gap: 24px; padding: 10px; border-bottom: 1px solid #f0f0f0; font-size: 13px; }}
  table {{ width: 100%; border-collapse: collapse; }}
  th {{ background: #fafafa; font-size: 12px; color: #888; font-weight: 500; padding: 10px 8px; text-align: left; border-bottom: 2px solid #f0f0f0; white-space: nowrap; }}
  td {{ padding: 10px 8px; font-size: 13px; border-bottom: 1px solid #f8f8f8; white-space: nowrap; }}
  td.name-col {{ white-space: normal; word-break: break-all; line-height: 1.4; max-width: 120px; }}
  tr:hover {{ background: #fafafa; }}
  .tag {{ display: inline-block; padding: 2px 6px; border-radius: 4px; font-size: 12px; font-weight: 600; }}
  .tag-up {{ background: #fde8e8; color: #e74c3c; }}
  .tag-down {{ background: #e8f8ed; color: #27ae60; }}
  .tag-flat {{ background: #f5f5f5; color: #999; }}
  .footer {{ padding: 12px 20px; font-size: 11px; color: #bbb; text-align: center; border-top: 1px solid #f0f0f0; }}
  .insight {{ padding: 14px 20px; background: #fffbf0; border-top: 1px solid #f0f0f0; }}
  .insight h3 {{ margin: 0 0 6px 0; font-size: 13px; color: #333; }}
  .insight p {{ margin: 3px 0; font-size: 12px; color: #666; line-height: 1.6; }}
</style>
</head>
<body>
<div class="container">
  <div class="header">
    <h1>📊 基金持仓日报</h1>
    <div class="sub">{gztime}</div>
  </div>
  <div class="overview">
    <div class="ov-item"><div class="num" style="color:{est_profit_color}">{total_est_profit:+,.2f}</div><div class="label">估算今日收益(元)</div></div>
    <div class="ov-item"><div class="num" style="color:#333">{total_amount:,.2f}</div><div class="label">总持有金额(元)</div></div>
  </div>
  <div class="count-bar">
    <span class="up">🟢 上涨 {up_count}</span>
    <span class="flat">⚪ 平盘 {flat_count}</span>
    <span class="down">🔴 下跌 {down_count}</span>
  </div>
  <table>
    <tr><th>基金名称</th><th>持有金额</th><th>涨跌幅</th><th>估算收益</th></tr>
"""

for r in results:
    if r["gszzl"] > 0:
        tag_class, num_class = "tag-up", "up"
    elif r["gszzl"] < 0:
        tag_class, num_class = "tag-down", "down"
    else:
        tag_class, num_class = "tag-flat", "flat"

    ep_color = "#e74c3c" if r["est_profit"] > 0 else "#27ae60" if r["est_profit"] < 0 else "#999"
    amount_str = f"{r['amount']:,.2f}" if r['amount'] > 0 else "-"

    # 基金名称每6个字换行
    name_display = r['name']
    name_parts = [name_display[i:i+6] for i in range(0, len(name_display), 6)]
    name_html = '<br>'.join(name_parts)

    html += f"""    <tr>
      <td class="name-col">{name_html}</td>
      <td>{amount_str}</td>
      <td><span class="tag {tag_class}"><span class="{num_class}">{r['gszzl_str']}</span></span></td>
      <td style="color:{ep_color};font-weight:600">{r['est_profit_str']}</td>
    </tr>
"""

html += f"""  </table>
  <div class="insight">
    <h3>📈 盘面速览</h3>
    <p>🔴 <b>涨幅最大</b>：{top['name']}（{top['gszzl_str']}）</p>
    <p>🟢 <b>跌幅最大</b>：{bottom['name']}（{bottom['gszzl_str']}）</p>
    <p>💰 <b>收益贡献最大</b>：{top_profit['name']}（{top_profit['est_profit_str']}元）</p>
    <p>📉 <b>亏损最大</b>：{bottom_profit['name']}（{bottom_profit['est_profit_str']}元）</p>
  </div>
  <div class="footer">⚠️ 估算收益 = 持有金额 × 涨跌幅，仅供参考，实际以基金公司披露为准</div>
</div>
</body>
</html>
"""

# ===== 发送邮件 =====
msg = MIMEMultipart("alternative")
msg["Subject"] = f"📊 基金日报 {today} | 估算收益 {total_est_profit:+,.0f}元"
msg["From"] = email_cfg["sender"]
msg["To"] = email_cfg["receiver"]

msg.attach(MIMEText(html, "html", "utf-8"))

try:
    server = smtplib.SMTP_SSL(email_cfg["smtp_server"], email_cfg["smtp_port"])
    server.login(email_cfg["sender"], email_cfg["password"])
    server.sendmail(email_cfg["sender"], email_cfg["receiver"], msg.as_string())
    server.quit()
    logging.info(f"✅ 邮件发送成功！估算收益: {total_est_profit:+,.2f}")
    print("✅ 邮件发送成功！")
except Exception as e:
    logging.error(f"❌ 邮件发送失败: {e}")
    print(f"❌ 邮件发送失败: {e}")
