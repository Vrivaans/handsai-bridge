#!/usr/bin/env node
"use strict";
/**
 * MCP Bridge para HandsAI
 * Conecta Claude Desktop con HandsAI Spring Boot server via HTTP
 */
Object.defineProperty(exports, "__esModule", { value: true });
const readline = require("readline");
const node_fetch_1 = require("node-fetch");
class HandsAIMcpBridge {
    constructor(handsaiUrl = 'http://localhost:8080') {
        this.handsaiBaseUrl = handsaiUrl;
        this.rl = readline.createInterface({
            input: process.stdin,
            output: process.stdout,
            terminal: false
        });
        // Solo logs de debug van a stderr
        console.error('🚀 HandsAI MCP Bridge iniciando...');
        console.error(`📡 Conectando a: ${this.handsaiBaseUrl}`);
    }
    async start() {
        this.rl.on('line', async (line) => {
            try {
                const request = JSON.parse(line.trim());
                const response = await this.handleMcpRequest(request);
                if (response) {
                    // Solo JSON puro a stdout
                    console.log(JSON.stringify(response));
                }
            }
            catch (error) {
                console.error(`❌ Error processing request: ${error}`);
                const errorResponse = {
                    jsonrpc: '2.0',
                    error: {
                        code: -32700,
                        message: 'Parse error'
                    }
                };
                console.log(JSON.stringify(errorResponse));
            }
        });
        this.rl.on('close', () => {
            console.error('👋 HandsAI MCP Bridge cerrando...');
            process.exit(0);
        });
    }
    async handleMcpRequest(request) {
        console.error(`🔧 Método: ${request.method}, ID: ${request.id}`);
        switch (request.method) {
            case 'initialize':
                return this.handleInitialize(request);
            case 'list_tools':
                return await this.handleListTools(request);
            case 'call_tool':
                return await this.handleCallTool(request);
            default:
                return {
                    jsonrpc: '2.0',
                    id: request.id,
                    error: {
                        code: -32601,
                        message: 'Method not found'
                    }
                };
        }
    }
    handleInitialize(request) {
        console.error('🔧 Initialize');
        return {
            jsonrpc: '2.0',
            id: request.id,
            result: {
                protocolVersion: '2024-11-05',
                capabilities: {
                    tools: {
                        listChanged: false
                    }
                },
                serverInfo: {
                    name: 'HandsAI MCP Bridge',
                    version: '1.0.0'
                }
            }
        };
    }
    async handleListTools(request) {
        console.error('📋 List tools - llamando a HandsAI API');
        try {
            const response = await (0, node_fetch_1.default)(`${this.handsaiBaseUrl}/mcp/tools/list`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            console.error(`✅ HandsAI respondió con ${data.result?.tools?.length || 0} herramientas`);
            // Convertir respuesta de HandsAI a formato MCP estándar
            const mcpTools = data.result?.tools?.map((tool) => ({
                name: tool.name,
                description: tool.description,
                inputSchema: tool.inputSchema || {
                    type: 'object',
                    properties: {},
                    required: []
                }
            })) || [];
            return {
                jsonrpc: '2.0',
                id: request.id,
                result: {
                    tools: mcpTools
                }
            };
        }
        catch (error) {
            console.error(`❌ Error llamando HandsAI list tools: ${error}`);
            return {
                jsonrpc: '2.0',
                id: request.id,
                error: {
                    code: -32603,
                    message: `Internal error: ${error instanceof Error ? error.message : String(error)}`
                }
            };
        }
    }
    async handleCallTool(request) {
        if (!request.params?.name) {
            return {
                jsonrpc: '2.0',
                id: request.id,
                error: {
                    code: -32602,
                    message: 'Invalid params: tool name required'
                }
            };
        }
        const toolName = request.params.name;
        const arguments_ = request.params.arguments || {};
        console.error(`🔧 Call tool: ${toolName} con args:`, arguments_);
        try {
            const mcpCallRequest = {
                jsonrpc: '2.0',
                id: 1,
                method: 'tools/call',
                params: {
                    name: toolName,
                    arguments: arguments_
                }
            };
            const response = await (0, node_fetch_1.default)(`${this.handsaiBaseUrl}/mcp/tools/call`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(mcpCallRequest)
            });
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            console.error(`✅ HandsAI ejecutó ${toolName} exitosamente`);
            // La respuesta de HandsAI ya debería estar en formato MCP correcto
            return {
                jsonrpc: '2.0',
                id: request.id,
                result: data.result || {
                    content: [{
                            type: 'text',
                            text: 'No result returned'
                        }]
                }
            };
        }
        catch (error) {
            console.error(`❌ Error ejecutando ${toolName}: ${error}`);
            return {
                jsonrpc: '2.0',
                id: request.id,
                error: {
                    code: -32603,
                    message: `Execution error: ${error instanceof Error ? error.message : String(error)}`
                }
            };
        }
    }
}
// Inicializar bridge
const bridge = new HandsAIMcpBridge();
bridge.start().catch((error) => {
    console.error(`💥 Error fatal: ${error}`);
    process.exit(1);
});
