# TuReparto
---

# Documento de Arquitectura: Ecosistema TuReparto & Tag-Mule

> **Versión:** 1.1 (Incluye Dashboard de Visualización)
> **Rol:** Arquitecto del Sistema
> **Estado:** Plan maestro previo al desarrollo.

---

## 1. Premisa y Restricciones

- **Tag-mule como caja negra:** El servicio de orquestación de IA (`tag-mule`) ya existe y es funcional en la red Docker. No se aborda su desarrollo aquí.
- **El servidor de TuReparto es intocable:** El código original que recibe los webhooks de Meta Cloud API y guarda en `tureparto.db` no se modifica bajo ninguna circunstancia.
- **Desacoplamiento total:** La inteligencia artificial es un addon externo. Si el sistema de IA se cae, el servidor de WhatsApp sigue funcionando y guardando mensajes perfectamente.

---

## 2. Arquitectura del Flujo de Datos

```
                  EXTERNO                         RED INTERNA DOCKER
┌───────────────┐      ┌──────────────┐      ┌────────────────────┐
│  Meta Cloud   │─────►│  TuReparto   │─────►│  tureparto.db     │
│  API (WSP)    │      │  Server :3000│      │  (Solo escritura  │
└───────────────┘      └──────────────┘      │  del server WSP)  │
                                             └────────┬──────────┘
                                                      │ (Lectura)
                                                      ▼
                                             ┌────────────────────┐
                                             │ cliente-tureparto  │
                                             │   (El Puente)      │
                                             │                    │
                                             │ [Motor 1: Poller]  │──── POST /enrich ────► tag-mule
                                             │                    │◄── 200 OK (job_id) ───
                                             │                    │
                                             │ [Motor 2: Webhook] │◄── POST /webhook ──── tag-mule
                                             │        :3001       │──── 200 OK ──────────►
                                             │                    │
                                             │ [Motor 3: Visor]   │──── GET /api/data ────► (Navegador)
                                             └────────┬───────────┘
                                                      │ (Escritura)
                                                      ▼
                                             ┌────────────────────┐
                                             │ tureparto_rich.db  │
                                             │  (Datos enriqueci- │
                                             │   dos por la IA)   │
                                             └────────────────────┘
```

---

## 3. Diseño de Bases de Datos

### 3.1 `tureparto.db` (Origen de datos - Intocable)
*Propiedad del servidor original.* 
El `cliente-tureparto` abre este archivo estrictamente en modo **Read-Only**. Solo lee la tabla de mensajes crudos (asumimos campos: `id`, `from_number`, `message_body`, `received_at`).

### 3.2 `tureparto_rich.db` (Destino de datos - Nueva)
*Propiedad exclusiva de `cliente-tureparto`*. Se crea desde cero para aislar la inteligencia artificial. 

```sql
CREATE TABLE IF NOT EXISTS ai_enrichment (
    original_msg_id TEXT PRIMARY KEY,   -- Relación 1:1 con tureparto.db
    phone_number TEXT NOT NULL,
    message_body TEXT NOT NULL,         -- Copia local para consultas fáciles
    status TEXT DEFAULT 'pending',      -- pending | sent_to_mule | completed | failed
    tag_mule_job_id TEXT,               -- Trazabilidad en caso de error
    suggested_tags TEXT,                -- JSON array: '["pedido", "bidones"]'
    error_details TEXT,                 -- Motivo del fallo si tag-mule devuelve error
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ai_status ON ai_enrichment(status);
```

---

## 4. Diseño del Componente: `cliente-tureparto`

Un binario de Go sin dependencias externas que corre como contenedor. Opera mediante 3 motores internos (goroutines).

### Motor 1: El Poller (Ingesta periódica)
- Se ejecuta en un *ticker* (ej. cada 30 segundos).
- Consulta los últimos `X` mensajes en `tureparto.db` (Read-Only).
- Por cada mensaje, verifica si `original_msg_id` ya existe en `tureparto_rich.db`.
- **Si no existe (es nuevo):**
  1. Lo inserta en `tureparto_rich.db` con `status='pending'`.
  2. Arma el payload y hace `POST` a `http://tag-mule:8080/api/v1/enrich`.
  3. Si tag-mule responde `200 OK`, actualiza el estado local a `sent_to_mule` y guarda el `job_id` devuelto en `tag_mule_job_id`.

### Motor 2: El Webhook (Receptor de resultados)
- Levanta un servidor HTTP interno.
- Escucha `POST /webhook` (el endpoint que tag-mule usa como callback en la configuración propuesta).
- Recibe el JSON con el resultado, busca el `original_msg_id` en `tureparto_rich.db` y actualiza a `status='completed'` guardando las `suggested_tags` y `tag_mule_job_id` si está presente.
- Responde `200 OK` a tag-mule.

### Motor 3: El Visor (Dashboard de Debug)
**Decisión de diseño:** Para monitorear el estado de la IA sin agregar complejidad (React, Node, bases de datos extra), el mismo binario de Go sirve una interfaz estática ultra liviana ("Pobre[...]")
- **`GET /`**: Sirve un archivo `index.html` estático (embebido o en carpeta `/static`). Fondo negro, fuente monoespaciada.
- **`GET /api/data`**: Endpoint que hace un `SELECT` a `tureparto_rich.db` (los últimos 50 registros ordenados por fecha) y devuelve un JSON crudo.
- **Flujo del Frontend:** El HTML contiene 10 líneas de JavaScript que hacen `fetch('/api/data')` y pintan el JSON en la pantalla. Se puede configurar un auto-recargo cada 30 segundos.
- **Propósito:** Permitir al administrador abrir `http://localhost:3001` en el navegador y ver inmediatamente si los mensajes de WhatsApp están pasando por la cola, si se están etiquetando y q[...]

---

## 5. Configuración Externa Requerida

En el archivo `config.yaml` del servicio `tag-mule`, el callback de TuReparto debe apuntar al nuevo contenedor cliente, no al servidor original:

```yaml
sources:
  tureparto:
    # Apunta al Motor 2 del cliente, no al server de WSP
    callback_url: http://cliente-tureparto:3001/webhook 
    # ... resto de la config (model, prompt, etc)
```

---

## 6. Integración Docker (`docker-compose.yml`)

La seguridad de no romper el servidor original se garantiza a nivel de sistema de archivos montando el volumen como `:ro` (Read-Only) para el cliente.

```yaml
services:
  tureparto-server:
    build: ./tureparto
    container_name: tureparto_server
    ports:
      - "3000:3000" # Expuesto para Tunnel de Meta
    volumes:
      - wsp_data:/app/data 
    networks:
      - backend_net

  cliente-tureparto:
    build: ./cliente_tureparto
    container_name: cliente_tureparto
    environment:
      - DB_SOURCE_PATH=/data/readonly/tureparto.db
      - DB_RICH_PATH=/data/rich/tureparto_rich.db
      - TAG_MULE_URL=http://tag-mule:8080/api/v1/enrich
      - POLL_INTERVAL_SECONDS=30
      - LISTEN_PORT=3001
    ports:
      - "3001:3001" # Expuesto SOLO para que el admin vea el Dashboard
    volumes:
      # Mismo volumen físico, pero montado como Solo Lectura
      - wsp_data:/data/readonly:ro 
      # Volumen propio e independiente para su base de datos enriquecida
      - rich_data:/data/rich
    depends_on:
      - tureparto-server
      - tag-mule
    networks:
      - backend_net

volumes:
  wsp_data:
  rich_data:

networks:
  backend_net:
```

---

## 7. Mapping de IDs y Endpoints (Resumen operativo)

- Callback que debe configurar `tag-mule` para este despliegue:

```yaml
sources:
  tureparto:
    callback_url: http://cliente-tureparto:3001/webhook
```

- Endpoint expuesto por `cliente-tureparto`:
  - POST http://cliente-tureparto:3001/webhook
  - Payload esperado (idéntico al contrato de tag-mule):

```json
{
  "job_id": "job-uuid-123",
  "item_id": "original-msg-id-456",
  "status": "completed",
  "tags": ["pedido", "producto"]
}
```

- Nota sobre IDs: `item_id` enviado a tag-mule por el `cliente-tureparto` corresponde a `original_msg_id` en `tureparto_rich.db`. En este documento usamos `original_msg_id` internamente; externamente (al llamar a tag-mule) lo enviamos como `item_id`.

- Acciones que hace el webhook receptor (`POST /webhook`):
  1. Buscar `original_msg_id` == `item_id` en `tureparto_rich.db`.
  2. Guardar `suggested_tags` (campo `tags` del payload) y `tag_mule_job_id` si viene `job_id`.
  3. Responder 200 OK.

---

## 8. Plan de Acción (Roadmap de Implementación)

1. **Esqueleto Go:** Crear estructura de carpetas, structs de BD y lógica de conexión a ambas bases de datos (una RO, otra RW).
2. **Motor 3 (Visor):** Crear el endpoint `GET /api/data` y el HTML estático. *Razón: Hacer esto primero permite probar la conexión a la BD rica inmediatamente desde el navegador.*
3. **Motor 2 (Webhook):** Levantar el `POST /webhook` y probar la actualización de registros simulando ser `tag-mule` (con curl).
4. **Motor 1 (Poller):** Implementar el *ticker*, la lectura de la BD original y la llamada HTTP a `tag-mule`.
5. **Dockerización:** Crear el `Dockerfile`, agregar al `docker-compose` general y hacer la prueba de fuego end-to-end con un mensaje real de WhatsApp.
