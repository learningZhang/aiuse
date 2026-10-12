# 统一封装好的MCP client
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client
from mcp.client.sse import sse_client
from mcp.client.streamable_http import streamablehttp_client
import asyncio
import sys
from abc import ABC,abstractmethod
from openai import OpenAI
from typing import Optional
from contextlib import AsyncExitStack
import json
class MCPClient:
    """
    MCP客户端
    """
    def __init__(self):
        self.session = None
        self.exit_stack = AsyncExitStack()        
    
    async def connect_stdio(self, script_path:str):
        """
        通过stdio的方式连接MCP
        """
        # 使用python启动，此时MCP Client和MCP Server使用同一套虚拟环境
        # server_params = StdioServerParameters(
        #     command = "python",
        #     args=[
        #         f"{script_path}"
        #     ],
        #     env=None
        # )

        # 使用当前 Python 解释器启动，Client 和 Server 使用同一套虚拟环境
        server_params = StdioServerParameters(
            command=sys.executable,
            args=[f"{script_path}"],
        )

        # 连接stdio mcp server
        stdio_transport = await self.exit_stack.enter_async_context(stdio_client(server_params))
        self.stdio, self.write = stdio_transport
        self.session = await self.exit_stack.enter_async_context(ClientSession(self.stdio, self.write))
        await self.session.initialize()
        
        print("✅ 成功连接到MCP服务器")
        await self.log("MCP Server可用服务器")
        tools = await self.list_tools()
        print(tools)
    async def connect_streamable_http(self, url:str):
        '''
        连接streamable http服务器
        Args:
            url: streamable http服务器地址，形如http://localhost:5001/mcp，请注意一般来说/mcp这个路径不用变，这是streamable-http默认路径
            更多的挂在方式请看官网https://github.com/modelcontextprotocol/python-sdk/blob/main/examples/snippets/servers/streamable_http_multiple_servers.py
        '''
        # 连接streamable mcp server
        stdio_transport = await self.exit_stack.enter_async_context(streamablehttp_client(url))
        readstream, writestream , _ = stdio_transport
        self.session = await self.exit_stack.enter_async_context(ClientSession(readstream, writestream))
        await self.session.initialize()
        
        print("✅ 成功连接到MCP服务器")
        await self.log("MCP Server可用服务器")
        tools = await self.list_tools()
        print(tools)
    async def list_tools(self):
        """
        列出可用工具,返回一个列表[Tool()，Tool()]
        """
        response = await self.session.list_tools()
        return response.tools
    
    async def call_tool(self, tool_name:str, tool_args:dict):
        """
        调用工具返回一个可迭代对象，其中有用的只有content元素
        """
        response = await self.session.call_tool(tool_name, tool_args)
        return response
    async def close(self):
        try:
            await self.exit_stack.aclose()
        except BaseException as e:
            # mcp 1.x 与 anyio 在关闭 stdio/streamable 传输时存在已知的 cancel scope
            # 竞态（会抛出 RuntimeError 或取消当前任务导致 CancelledError），
            # 此时子进程通常已按 MCP 规范正常退出，捕获（含 CancelledError）避免程序退出时崩溃
            print(f"关闭连接时出现可忽略的异常: {type(e).__name__}: {e}")
        
    async def log(self, message:str):
        print(f"================================={message}=================================")

class MCPAgent:
    """
    支持MCP协议的LLM
    """

    # 单个工具结果写入 history 的最大字符数，防止撑爆上下文窗口
    MAX_TOOL_RESULT_CHARS = 2000

    # 终端打印工具结果的最大字符数，仅影响显示
    MAX_TOOL_PRINT_CHARS = 500

    # 单次提问内最多允许的工具调用轮次，防止模型陷入死循环
    MAX_TOOL_ITERATIONS = 5

    def __init__(self, mcp_client_list:list[MCPClient] = []):
        self.model = OpenAI(
            base_url='http://localhost:11434/v1',
            api_key='ollama', # 本地 Ollama 不校验，占位即可
        )
        self.mcp_client_list = mcp_client_list
        
        # 聊天记录
        self.history = [
            # {"role": "system", "content": "你是一个智能问答助手，你可以使用用户给出的工具回答用户的问题"}
        ] 

    @staticmethod
    def truncate_text(text, max_chars:int):
        """
        截断过长的文本，并在末尾标注被省略的字符数
        """
        if text is None:
            return ""

        if not isinstance(text, str):
            text = str(text)

        if len(text) <= max_chars:
            return text

        omitted = len(text) - max_chars
        return f"{text[:max_chars]}\n...[已截断 {omitted} 字符，共 {len(text)} 字符]"

    async def get_mcp_tools(self):
        """
        获取MCP工具列表
        """
        available_tools = []
        for mcp_id, client in enumerate(self.mcp_client_list):
            tools = await client.list_tools()
            for tool in tools:
                available_tools.append(
                    (
                        mcp_id,
                        {
                            "name": tool.name+f"_id_{mcp_id}",
                            "description": tool.description,
                            "input_schema": tool.inputSchema    
                        }
                    )
                )
        return available_tools
    
    async def get_tool2mcp_dict(self):
        """
        获取工具名称到MCP客户端的映射
        """
        tool2mcp_dict = {}
        tool2mcp_origin_name = {}

        for mcp_id, client in enumerate(self.mcp_client_list):
            tools = await client.list_tools()
            for tool in tools:
                tool2mcp_dict[tool.name+f"_id_{mcp_id}"] = mcp_id
                tool2mcp_origin_name[tool.name+f"_id_{mcp_id}"] = tool.name
        return tool2mcp_dict, tool2mcp_origin_name
    async def convert2openaiformat(self, available_tools:list):
        """
        将MCP工具转换为OpenAI格式的tool列表
        """
        openai_tools = []
        for mcp_id, tool_info in available_tools:
            # 构建OpenAI格式的工具定义
            openai_tool = {
                "type": "function",
                "function": {
                    "name": tool_info["name"],
                    "description": tool_info["description"],
                    "parameters": tool_info["input_schema"]
                }
            }
            openai_tools.append(openai_tool)
        return openai_tools

        
    async def process_query(self, query:str):
        """
        回答用户问题
        """
        self.history.append({"role": "user", "content": query})

        # 获取工具列表
        available_tools = await self.get_mcp_tools()
        tool2mcp_dict, tool2mcp_origin_name = await self.get_tool2mcp_dict()
        
        # 将工具转为openai格式
        openai_tool_list = await self.convert2openaiformat(available_tools)


        # 调用大模型
        response = self.model.chat.completions.create(
            model='qwen3:14b', # 本地 Ollama Model-Id
            messages=self.history,
            tools=openai_tool_list
        )
    
        final_text = []

        # 处理大模型的回复
        iteration = 0
        while True:
            assistant_messages = json.loads(response.choices[0].message.to_json())

            # 注意：OpenAI 官方在无工具调用时返回 "tool_calls": null，
            # 而 Ollama 等兼容实现会直接省略该字段，故用 get 兼容三种情况：
            # 键缺失 / 值为 None / 空列表
            tool_calls = assistant_messages.get("tool_calls")
            assistant_content = assistant_messages.get("content") or ""

            if not tool_calls:
                # 没有工具调用直接停止
                final_text.append(f"Assitant: {assistant_content}")
                print(f"Assitant: {assistant_content}")
                self.history.append(assistant_messages)
                break
            else:
                # 达到工具调用轮次上限，停止继续调用工具
                if iteration >= self.MAX_TOOL_ITERATIONS:
                    warning = f"已达到工具调用上限（{self.MAX_TOOL_ITERATIONS} 轮），停止本次工具调用。"
                    final_text.append(f"Assitant: {warning}")
                    print(f"Assitant: {warning}")
                    break

                iteration += 1

                # 有工具调用，需要调用MCP Server
                if assistant_content != "":
                    final_text.append(f"Assitant: {assistant_content}")
                    print(f"========Assitant: {assistant_content}")  
                                
                # 调用工具
                tool_info = tool_calls[0]
                tool_name = tool_info['function']['name']
                print(f"========Tool Calls: 调用工具{tool_name}")
                final_text.append(f"Tool Calls: 调用工具{tool_name}")
                tool_args = json.loads(tool_info['function']['arguments'])
                mcp_id = tool2mcp_dict[tool_name]
                mcp_client = self.mcp_client_list[mcp_id]
                tool_response = await mcp_client.call_tool(tool2mcp_origin_name[tool_name], tool_args)
                raw_tool_result_text = tool_response.content[0].text

                # 截断后再写入history，避免超长工具结果撑爆上下文
                tool_result_text = self.truncate_text(raw_tool_result_text, self.MAX_TOOL_RESULT_CHARS)
                
                # 格式化工具调用结果
                final_text.append(f"Tool Response: {tool_result_text}")
                print(f"========Tool Response: {self.truncate_text(tool_result_text, self.MAX_TOOL_PRINT_CHARS)}")

                # 大模型的回复
                self.history.append(assistant_messages)
                
                # 工具调用的结果
                self.history.append({
                    "role": "tool",
                    "content": tool_result_text,
                    "tool_call_id": tool_info['id']
                })

                # 调用大模型继续回复
                response = self.model.chat.completions.create(
                    model='qwen3:14b', # 本地 Ollama Model-Id
                    messages=self.history,
                    tools=openai_tool_list
                )
        
        return final_text

    async def chat(self):
        """
        回答用户问题
        """
        print("\nMCP Agent Started!")
        print("Type your queries or 'quit' to exit.")

        while True:
            try:
                query = (await asyncio.to_thread(input, "\nUser: ")).strip()

                if query.lower() == 'quit':
                    break
                await self.process_query(query)

            except Exception as e:
                print(f"\nError: {str(e)}")

async def main():
    # stdio通讯
    mcp_server_list = [
        "./bilibili_mcp.py"
    ]
    
    # streamable http通讯
    mcp_url_list = [
        "http://localhost:5001/mcp",  # weather_mcp.py
        "http://localhost:5002/mcp",  # time_mcp.py
    ]

    mcp_client_list = []
    failed_servers = []
    try:
        # 连接MCP Server
        
        # stdio服务器
        for mcp_server_name in mcp_server_list:
            print(f"=======================MCP Server {mcp_server_name} 初始化=======================")
            client = MCPClient()
            try:
                await client.connect_stdio(mcp_server_name)
                mcp_client_list.append(client)
            except Exception as e:
                # 单个Server连接失败不影响其他Server，记录后继续
                failed_servers.append((mcp_server_name, f"{type(e).__name__}: {e}"))
                print(f"❌ 连接失败 {mcp_server_name}: {type(e).__name__}: {e}")
                await client.close()

        # streamable服务器
        for mcp_url in mcp_url_list:
            print(f"=======================MCP Server {mcp_url} 初始化=======================")
            client = MCPClient()
            try:
                await client.connect_streamable_http(mcp_url)
                mcp_client_list.append(client)
            except Exception as e:
                # 常见原因：目标Server未启动，或端口/路径不正确
                failed_servers.append((mcp_url, f"{type(e).__name__}: {e}"))
                print(f"❌ 连接失败 {mcp_url}: {type(e).__name__}: {e}")
                print("   请确认对应的MCP Server已启动，且端口与路径正确。")
                await client.close()

        # 连接结果汇总
        print(f"\n连接成功 {len(mcp_client_list)} 个，失败 {len(failed_servers)} 个")
        for name, reason in failed_servers:
            print(f"  - 失败: {name} ({reason})")

        if not mcp_client_list:
            print("\n没有任何MCP Server连接成功，退出。请先启动Server后重试：")
            print("  venv/bin/python3 weather_mcp.py   # :5001")
            print("  venv/bin/python3 time_mcp.py      # :5002")
            return

        # 初始化MCP Agent
        mcp_agent = MCPAgent(mcp_client_list)
        await mcp_agent.chat()
    finally:
        for client in mcp_client_list:
            await client.close()


if __name__ == "__main__":
    asyncio.run(main())