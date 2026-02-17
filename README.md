# HandsAI Bridge

Este proyecto implementa un servidor MCP (Model Context Protocol) en Node.js, diseñado para actuar como un puente de comunicación para el proyecto [HandsAI](https://github.com/Vrivaans/handsai). Permite que el servidor principal de Java de HandsAI se conecte e interactúe con este bridge a través de un socket.

## Características

- Implementación de un servidor MCP utilizando Node.js y TypeScript.
- Configuración sencilla para la integración con clientes MCP.
- Diseñado para ser ligero y fácil de extender.

## Requisitos Previos

- [Node.js](https://nodejs.org/) (versión 20.x o superior)
- [npm](https://www.npmjs.com/) (generalmente se instala con Node.js)

## Instalación

1. Clona este repositorio en tu máquina local:
   ```bash
   git clone https://github.com/Vrivaans/handsai-bridge.git
   cd handsai-bridge
   ```

2. Instala las dependencias del proyecto. Este proyecto no tiene dependencias de producción, pero sí de desarrollo como TypeScript.
   ```bash
   npm install
   ```

## Uso

Este proyecto está diseñado para ser ejecutado como un servidor MCP. El cliente MCP (como el que se usa en el proyecto HandsAI) se configurará para conectarse a este servidor.

### Configuración del Cliente MCP

Para que un cliente MCP se conecte a este bridge, debes configurarlo de manera similar al siguiente ejemplo. Este JSON le indica al cliente cómo iniciar y comunicarse con el servidor de `handsai-bridge`.

Asegúrate de reemplazar la ruta en `args` con la ruta **absoluta** al archivo `index.ts` en tu sistema.

```json
{
  "mcpServers": {
    "handsai": {
      "command": "npx",
      "args": [
        "-y",
        "tsx",
        "/ruta/absoluta/a/tu/proyecto/handsai-bridge/index.ts"
      ]
    }
  }
}
```

> **Nota:** Si experimentas errores de permisos (`EPERM`) al conectar con Claude Desktop en macOS, intenta mover la carpeta del proyecto a una ubicación pública como `Documentos`. MacOS a veces restringe el acceso de aplicaciones a carpetas como `Escritorio`, `Descargas` o carpetas sincronizadas.

### Ejecución

Una vez que el cliente MCP esté configurado, iniciará automáticamente el servidor `handsai-bridge` cuando sea necesario, utilizando el comando y los argumentos especificados. No necesitas ejecutar el servidor manualmente.

El archivo `index.ts` contiene la lógica principal para levantar el servidor, escuchar conexiones y manejar la comunicación entre el cliente y el bridge.

## Conexión con HandsAI

El propósito principal de este bridge es servir como un punto de conexión para el servidor de Java del proyecto [HandsAI](https://github.com/Vrivaans/handsai). El servidor de Java actuará como cliente de este bridge, permitiendo el intercambio de información y comandos entre ambos sistemas.

## Configuración Avanzada

Por defecto, el bridge intenta conectarse a `http://localhost:8080`.

Al iniciar por primera vez, el servidor **creará automáticamente** un archivo `config.json` en la misma carpeta que `index.ts` con la configuración por defecto. Puedes editar este archivo para cambiar la URL:

```json
{
  "handsaiUrl": "http://tu-servidor:puerto"
}
```
