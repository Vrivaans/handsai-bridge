#!/usr/bin/env node
/**
 * MCP Bridge para HandsAI
 * Conecta MCP clients con HandsAI Spring Boot server via HTTP
 */
import * as readline from 'readline';
class HandsAIMcpBridge {
    handsaiBaseUrl;
    rl;
    constructor(handsaiUrl = 'http://localhost:8080') {
        this.handsaiBaseUrl = handsaiUrl;
        this.rl = readline.createInterface({
            input: process.stdin,
            output: process.stdout,
            terminal: false
        });
    }
    async start() {
        this.rl.on('line', async (line) => {
            try {
                const request = JSON.parse(line.trim());
                const response = await this.handleMcpRequest(request);
                if (response) {
                    console.log(JSON.stringify(response));
                }
            }
            catch (error) {
                // SIEMPRE incluir el ID de la request original, incluso en errores de parse
                const errorResponse = {
                    jsonrpc: '2.0',
                    id: null, // Para errores de parse, se usa null según JSON-RPC 2.0
                    error: {
                        code: -32700,
                        message: 'Parse error',
                        data: error instanceof Error ? error.message : String(error)
                    }
                };
                console.log(JSON.stringify(errorResponse));
            }
        });
        this.rl.on('close', () => {
            process.exit(0);
        });
    }
    async handleMcpRequest(request) {
        // CRÍTICO: Validar que el request tenga la estructura mínima
        if (!request.jsonrpc || request.jsonrpc !== '2.0') {
            return {
                jsonrpc: '2.0',
                id: request.id || null,
                error: {
                    code: -32600,
                    message: 'Invalid Request: jsonrpc must be "2.0"'
                }
            };
        }
        if (!request.method) {
            return {
                jsonrpc: '2.0',
                id: request.id || null,
                error: {
                    code: -32600,
                    message: 'Invalid Request: method is required'
                }
            };
        }
        switch (request.method) {
            case 'initialize':
                return this.handleInitialize(request);
            case 'tools/list':
                return await this.handleListTools(request);
            case 'tools/call':
                return await this.handleCallTool(request);
            case 'notifications/initialized':
                // Este es un notification, no requiere respuesta
                return null;
            default:
                return {
                    jsonrpc: '2.0',
                    id: request.id || null,
                    error: {
                        code: -32601,
                        message: `Method not found: ${request.method}`
                    }
                };
        }
    }
    handleInitialize(request) {
        return {
            jsonrpc: '2.0',
            id: request.id, // Usar el ID original de la request
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
        try {
            const response = await fetch(`${this.handsaiBaseUrl}/mcp/tools/list`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            // Validar que la respuesta tenga la estructura esperada
            if (!data || typeof data !== 'object') {
                throw new Error('Invalid response format from HandsAI server');
            }
            // Manejar diferentes formatos de respuesta del servidor HandsAI
            let tools = [];
            if (data.result && Array.isArray(data.result.tools)) {
                tools = data.result.tools;
            }
            else if (Array.isArray(data.tools)) {
                tools = data.tools;
            }
            else if (Array.isArray(data)) {
                tools = data;
            }
            // Convertir a formato MCP estándar
            const mcpTools = tools.map((tool) => ({
                name: tool.name || 'unknown',
                description: tool.description || 'No description available',
                inputSchema: tool.inputSchema || {
                    type: 'object',
                    properties: {},
                    required: []
                }
            }));
            return {
                jsonrpc: '2.0',
                id: request.id, // Usar el ID original
                result: {
                    tools: mcpTools
                }
            };
        }
        catch (error) {
            return {
                jsonrpc: '2.0',
                id: request.id, // Usar el ID original
                error: {
                    code: -32603,
                    message: 'Internal error',
                    data: error instanceof Error ? error.message : String(error)
                }
            };
        }
    }
    async handleCallTool(request) {
        // Validar parámetros requeridos
        if (!request.params || !request.params.name) {
            return {
                jsonrpc: '2.0',
                id: request.id, // Usar el ID original
                error: {
                    code: -32602,
                    message: 'Invalid params: tool name is required'
                }
            };
        }
        const toolName = request.params.name;
        const arguments_ = request.params.arguments || {};
        try {
            const mcpCallRequest = {
                jsonrpc: '2.0',
                id: `internal-${Date.now()}`, // ID único para evitar conflictos
                method: 'tools/call',
                params: {
                    name: toolName,
                    arguments: arguments_
                }
            };
            const response = await fetch(`${this.handsaiBaseUrl}/mcp/tools/call`, {
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
            // Validar respuesta del servidor
            if (!data || typeof data !== 'object') {
                throw new Error('Invalid response format from HandsAI server');
            }
            // Manejar diferentes formatos de respuesta
            let result;
            if (data.result) {
                result = data.result;
            }
            else if (data.content) {
                result = { content: data.content };
            }
            else {
                result = {
                    content: [{
                        type: 'text',
                        text: JSON.stringify(data)
                    }]
                };
            }
            return {
                jsonrpc: '2.0',
                id: request.id, // CRÍTICO: Usar el ID original de la request
                result: result
            };
        }
        catch (error) {
            return {
                jsonrpc: '2.0',
                id: request.id, // Usar el ID original
                error: {
                    code: -32603,
                    message: 'Execution error',
                    data: error instanceof Error ? error.message : String(error)
                }
            };
        }
    }
}
// Inicializar bridge
const bridge = new HandsAIMcpBridge();
bridge.start().catch((error) => {
    console.error('Bridge startup error:', error);
    process.exit(1);
});
