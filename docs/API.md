# 📡 API Reference — TuReparto Webhook

Documentación técnica de los endpoints del servidor de webhook de WhatsApp.

---

## 📍 Base URL

```
http://localhost:3000
```

Cuando usas localtunnel o Cloudflare Tunnel, reemplaza con tu URL pública:
```
https://xxxx.loca.lt
```

---

## `GET /`

Página de estado del servidor.

### Respuesta

```
Status: 200 OK
Content-Type: text/plain; charset=utf-8

✅ Servidor de WhatsApp Webhook funcionando!
📌 Endpoint: /webhook
```

---

## `GET /webhook`

Endpoint de verificación utilizado por Meta WhatsApp Cloud API para confirmar que el webhook es válido.

### Query Parameters

| Parámetro | Tipo | Requerido | Descripción |
|-----------|------|-----------|-------------|
| `hub.mode` | `string` | ✅ | Debe ser `"subscribe"` |
| `hub.verify_token` | `string` | ✅ | Token secreto configurado en `VERIFY_TOKEN` |
| `hub.challenge` | `string` | ✅ | String que Meta espera que devuelvas para confirmar |

### Ejemplo

```bash
curl -s -G 'http://localhost:3000/webhook' \
  --data-urlencode 'hub.mode=subscribe' \
  --data-urlencode 'hub.verify_token=tu_token_seguro_aqui' \
  --data-urlencode 'hub.challenge=987654321'
```

### Respuestas

| Código | Condición | Body |
|--------|-----------|------|
| `200` | Token válido y mode=subscribe | El `challenge` recibido (ej: `987654321`) |
| `403` | Token inválido o mode incorrecto | `Token de verificación inválido` |

### Logs del servidor

**Éxito:**
```
📥 Verificación recibida: mode=subscribe, token=tu_token_seguro_aqui
✅ Verificación exitosa! Meta ha confirmado el webhook.
```

**Fallo:**
```
📥 Verificación recibida: mode=subscribe, token=token_incorrecto
❌ Verificación fallida: token incorrecto (esperado: tu_token_seguro_aqui)
```

---

## `POST /webhook`

Recibe mensajes entrantes de WhatsApp enviados por Meta y **los guarda automáticamente en SQLite**.

### Headers

| Header | Valor esperado |
|--------|----------------|
| `Content-Type` | `application/json` |

### Body (JSON)

El payload sigue la estructura de la [WhatsApp Cloud API Webhook](https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks/payloads).

```json
{
  "object": "whatsapp_business_account",
  "entry": [{
    "id": "WHATSAPP_BUSINESS_ACCOUNT_ID",
    "changes": [{
      "value": {
        "messaging_product": "whatsapp",
        "metadata": {
          "display_phone_number": "15551234567",
          "phone_number_id": "PHONE_NUMBER_ID"
        },
        "contacts": [{
          "profile": { "name": "Juan Pérez" },
          "wa_id": "5215512345678"
        }],
        "messages": [{
          "from": "5215512345678",
          "id": "wamid.XXX",
          "timestamp": "1700000000",
          "type": "text",
          "text": {
            "body": "Hola! Quiero hacer un pedido"
          }
        }]
      },
      "field": "messages"
    }]
  }]
}
```

### Lo que guarda en la BD

| Campo SQLite | Origen | Descripción |
|-------------|--------|-------------|
| `from_number` | `entry[].changes[].value.messages[].from` | Número de teléfono del remitente |
| `message_body` | `entry[].changes[].value.messages[].text.body` | Contenido del mensaje de texto |
| `received_at` | `CURRENT_TIMESTAMP` | Fecha/hora de recepción (automático) |

### Ejemplo

```bash
curl -s -X POST http://localhost:3000/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "entry": [{
      "changes": [{
        "value": {
          "messages": [{
            "from": "5215512345678",
            "text": { "body": "Hola! Quiero hacer un pedido" }
          }]
        }
      }]
    }]
  }'
```

### Respuestas

| Código | Body | Descripción |
|--------|------|-------------|
| `200` | `OK` | Mensaje recibido, procesado y **guardado en BD** |
| `400` | `Error leyendo request` | Body inválido o error de lectura |

### Logs del servidor

```log
📩 Webhook recibido:
{
  "Entry": [
    {
      "Changes": [
        {
          "Value": {
            "Messages": [
              {
                "From": "5215512345678",
                "Text": {
                  "Body": "Hola! Quiero hacer un pedido"
                }
              }
            ]
          }
        }
      ]
    }
  ]
}
💬 De: 5215512345678 | Mensaje: Hola! Quiero hacer un pedido
💾 Mensaje guardado en BD ✅
```

---

## Manejo de errores del servidor

| Código | Condición |
|--------|-----------|
| `404` | Ruta no encontrada (ej: `/otra-ruta`) |
| `405` | Método HTTP no permitido (ej: `PUT /webhook`) |
| `403` | Token de verificación inválido en GET /webhook |
| `400` | Error al leer el body del POST |

---

## Pruebas rápidas

### Script de prueba completo

```bash
#!/bin/bash
echo "=== 1. Health Check ==="
curl -s http://localhost:3000/
echo

echo "=== 2. Verificación ==="
curl -s -G 'http://localhost:3000/webhook' \
  --data-urlencode 'hub.mode=subscribe' \
  --data-urlencode 'hub.verify_token=tu_token_seguro_aqui' \
  --data-urlencode 'hub.challenge=123456'
echo

echo "=== 3. Mensaje de texto ==="
curl -s -X POST http://localhost:3000/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "entry": [{
      "changes": [{
        "value": {
          "messages": [{
            "from": "5215512345678",
            "text": { "body": "Hola, quiero información" }
          }]
        }
      }]
    }]
  }'
echo

echo "=== 4. Verificar que se guardó en BD ==="
sqlite3 tureparto.db "SELECT * FROM messages;"
echo

echo "=== 5. Token incorrecto (debe fallar) ==="
curl -s -G 'http://localhost:3000/webhook' \
  --data-urlencode 'hub.mode=subscribe' \
  --data-urlencode 'hub.verify_token=token_mal' \
  --data-urlencode 'hub.challenge=123456'
echo
```

---

## Tipos de mensajes soportados

Actualmente TuReparto procesa:

| Tipo | Descripción | Estado |
|------|-------------|--------|
| `text` | Mensajes de texto (se guardan en BD) | ✅ Soportado |
| `image` | Imágenes | 🔜 Próximamente |
| `document` | Documentos | 🔜 Próximamente |
| `audio` | Notas de voz | 🔜 Próximamente |
| `video` | Videos | 🔜 Próximamente |
| `location` | Ubicación | 🔜 Próximamente |
| `contacts` | Contactos | 🔜 Próximamente |
| `interactive` | Respuestas a botones | 🔜 Próximamente |
| `order` | Pedidos | 🔜 Próximamente |
