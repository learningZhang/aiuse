#!/usr/bin/env python3
"""
Bilibili MCP Server using FastMCP
提供B站视频搜索功能的MCP服务器
"""
import asyncio 
from typing import List, Dict, Any, Union, cast  
from mcp.server.fastmcp import FastMCP
from bilibili_api import search, sync
# 创建FastMCP实例
mcp = FastMCP("bilibili")
@mcp.tool()
async def search_articles(keyword: str, page: int = 1, page_size: int = 10):
    """
    根据关键词搜索B站文章
    
    Args:
        keyword: 搜索关键词
        page: 页码，默认为1
        page_size: 每页数量，默认为10
        
    Returns:
        List[Dict]: 文章信息列表，包含标题、UP主、播放量、链接等信息
    """
    try:
        # 使用bilibili-api进行搜索
        result: Dict[str, Any] = await search.search(keyword, page=page)        
        return result["result"]
        
    except Exception as e:
        return [{"error": f"搜索失败: {str(e)}"}]
    
@mcp.tool()
async def search_videos(keyword: str, page: int = 1, page_size: int = 10):  
    """
    根据关键词搜索B站视频
    
    Args:
        keyword: 搜索关键词
        page: 页码，默认为1
        page_size: 每页数量，默认为10
        
    Returns:
        List[Dict]: 视频信息列表，包含标题、UP主、播放量、链接等信息
    """
    try:
        # 使用bilibili-api进行搜索
        result: Dict[str, Any] = await search.search(keyword, page=page)
        videos = []
        for item in result["result"]:
            if item["result_type"] == "video":
                videos.append(item)
        return videos
        
    except Exception as e:
        return [{"error": f"搜索失败: {str(e)}"}]

if __name__ == "__main__":
    # 启动MCP服务器
    print("MCP server started")
    mcp.run(transport="stdio")
