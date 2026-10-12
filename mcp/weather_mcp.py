import requests
from mcp.server.fastmcp import FastMCP
import random
import datetime
from typing import Dict, Union
# import numpy as np
# 初始化mcp服务器
mcp = FastMCP("weather_mcp_server",port=5001)
@mcp.tool()
async def get_location_id(location_name: str) -> Dict[str, Union[str, int]]:
    """
    根据城市名称获取城市ID
    
    Args:
        location_name: 城市名称
        
    Returns:
        包含城市名称和城市ID的字典
    """
    # data = np.random.rand(10000, 2)
    return {
        "location_name": location_name,
        "location_id": random.randint(1, 9999)
    }
@mcp.tool()
async def get_weather(location_id: int, date: str) -> Dict[str, Union[int, str]]:
    """
    根据城市ID和日期获取天气信息
    
    Args:
        location_id: 城市ID
        date: 日期，格式为YYYY-MM-DD
        
    Returns:
        包含天气信息的字典，包括城市ID、日期、天气状况、温度和湿度
    """
    return {
        "location_id": location_id,
        "date": date,
        "weather": random.choice(["晴", "阴", "雨", "雪", "雾", "霾"]),
        "temperature": random.randint(-10, 35),
        "humidity": random.randint(0, 100)
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
if __name__ == "__main__":
    print("天气MCP服务器已启动")
    mcp.run(transport="streamable-http")