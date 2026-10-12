#!/usr/bin/env python3
"""
Time MCP Server using FastMCP
提供时间查询功能的MCP服务器
"""
import datetime
from zoneinfo import ZoneInfo, available_timezones
from typing import Dict, Union
from mcp.server.fastmcp import FastMCP

# 初始化mcp服务器
mcp = FastMCP("time_mcp_server", port=5002)

# 默认时区
DEFAULT_TIMEZONE = "Asia/Shanghai"


@mcp.tool()
async def get_current_time(timezone: str = DEFAULT_TIMEZONE) -> Dict[str, Union[str, int]]:
    """
    获取指定时区的当前时间

    Args:
        timezone: IANA时区名称，例如 Asia/Shanghai、UTC、America/New_York，默认为 Asia/Shanghai

    Returns:
        包含时区、日期、时间、星期和时间戳的字典
    """
    try:
        tz = ZoneInfo(timezone)
    except Exception:
        return {
            "error": f"无效的时区: {timezone}，请使用 IANA 时区名称，例如 Asia/Shanghai"
        }

    now = datetime.datetime.now(tz)
    weekday_names = ["星期一", "星期二", "星期三", "星期四", "星期五", "星期六", "星期日"]

    return {
        "timezone": timezone,
        "datetime": now.strftime("%Y-%m-%d %H:%M:%S"),
        "date": now.strftime("%Y-%m-%d"),
        "time": now.strftime("%H:%M:%S"),
        "weekday": weekday_names[now.weekday()],
        "utc_offset": now.strftime("%z"),
        "timestamp": int(now.timestamp())
    }


@mcp.tool()
async def get_current_utc_datetime() -> Dict[str, str]:
    """
    获取当前的UTC时间

    Returns:
        包含当前UTC时间的字典，时间格式为YYYY-MM-DD HH:MM:SS
    """
    return {
        "datetime": datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%d %H:%M:%S")
    }


@mcp.tool()
async def get_timestamp() -> Dict[str, int]:
    """
    获取当前的Unix时间戳

    Returns:
        包含秒级和毫秒级时间戳的字典
    """
    now = datetime.datetime.now(datetime.timezone.utc)
    return {
        "timestamp": int(now.timestamp()),
        "timestamp_ms": int(now.timestamp() * 1000)
    }


@mcp.tool()
async def convert_timezone(
    datetime_str: str,
    from_timezone: str,
    to_timezone: str
) -> Dict[str, str]:
    """
    将某个时区的时间转换到另一个时区

    Args:
        datetime_str: 时间字符串，格式为 YYYY-MM-DD HH:MM:SS
        from_timezone: 源时区的IANA名称，例如 Asia/Shanghai
        to_timezone: 目标时区的IANA名称，例如 America/New_York

    Returns:
        包含源时间和转换后时间的字典
    """
    try:
        from_tz = ZoneInfo(from_timezone)
        to_tz = ZoneInfo(to_timezone)
    except Exception:
        return {
            "error": f"无效的时区: {from_timezone} 或 {to_timezone}"
        }

    try:
        naive = datetime.datetime.strptime(datetime_str, "%Y-%m-%d %H:%M:%S")
    except ValueError:
        return {
            "error": f"时间格式错误: {datetime_str}，应为 YYYY-MM-DD HH:MM:SS"
        }

    source = naive.replace(tzinfo=from_tz)
    target = source.astimezone(to_tz)

    return {
        "source_timezone": from_timezone,
        "source_datetime": source.strftime("%Y-%m-%d %H:%M:%S"),
        "target_timezone": to_timezone,
        "target_datetime": target.strftime("%Y-%m-%d %H:%M:%S")
    }


@mcp.tool()
async def list_common_timezones() -> Dict[str, Union[list, int]]:
    """
    列出常用时区名称，便于选择合法的时区参数

    Returns:
        包含常用时区列表和系统可用时区总数的字典
    """
    common = [
        "UTC",
        "Asia/Shanghai",
        "Asia/Hong_Kong",
        "Asia/Taipei",
        "Asia/Tokyo",
        "Asia/Seoul",
        "Asia/Singapore",
        "Europe/London",
        "Europe/Paris",
        "Europe/Berlin",
        "America/New_York",
        "America/Chicago",
        "America/Los_Angeles",
        "Australia/Sydney"
    ]
    return {
        "common_timezones": common,
        "available_count": len(available_timezones())
    }


if __name__ == "__main__":
    print("时间MCP服务器已启动")
    mcp.run(transport="streamable-http")
