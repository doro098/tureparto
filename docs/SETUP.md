# 🔌 Guía de Integración con Meta WhatsApp Cloud API

Guía paso a paso para conectar TuReparto con la API de WhatsApp de Meta.

---

## 📋 Prerrequisitos

- [ ] Servidor TuReparto corriendo (`http://localhost:3000`)
- [ ] Cloudflare Tunnel activo (URL pública HTTPS)
- [ ] Cuenta en [Meta for Developers](https://developers.facebook.com/)
- [ ] Número de teléfono para WhatsApp Business

---

## 🧭 Paso 1: Crear una app en Meta Developers

1. Ve a [developers.facebook.com](https://developers.facebook.com/)
2. Inicia sesión con tu cuenta de Facebook
3. Haz clic en **"My Apps"** → **"Create App"**
4. Selecciona **"Business"** como tipo de app
5. Completa los datos:
   - **App Name:** `TuReparto` (o el nombre que quieras)
   - **Contact Email:** Tu correo electrónico
6. Haz clic en **"Create App ID"**

---

## 🧭 Paso 2: Agregar producto WhatsApp

1. Dentro de tu app, busca la sección **"Add Product"**
2. Localiza **WhatsApp** y haz clic en **"Set Up"**
3. Serás redirigido a la configuración de WhatsApp

---

## 🧭 Paso 3: Configurar el Webhook

### 3.1 Obtener URL pública

Elegí una de estas opciones para exponer tu servidor local a internet:

#### Opción A: localtunnel (más simple)

```bash
# Instalar
npm install -g localtunnel

# Exponer (en otra terminal)
lt --port 3000
```

Verás una salida como esta:

```
your url is: https://abeja-verde.loca.lt
```

> ⚠️ **Conserva esta URL.** La necesitarás en el siguiente paso.

#### Opción B: Cloudflare Tunnel

```bash
cloudflared tunnel --url http://localhost:3000
```

Verás una salida como esta:

```
2024-01-15T10:30:00Z INF |  https://abeja-verde.trycloudflare.com                                                    |
```

### 3.2 Configurar en Meta

1. En el panel de WhatsApp de tu app, ve a **"Configuration"**
2. En la sección **"Webhook"**, haz clic en **"Edit"**
3. Completa los campos:

   | Campo | Valor |
   |-------|-------|
   | **Callback URL** | `https://abeja-verde.trycloudflare.com/webhook` |
   | **Verify Token** | `tu_token_seguro_aqui` |

   > El **Verify Token** debe ser exactamente el mismo que configuraste al iniciar TuReparto con `export VERIFY_TOKEN="tu_token_seguro_aqui"`.

4. Haz clic en **"Verify and Save"**

### 3.3 Verificación exitosa

Si todo funciona, Meta mostrará ✅ **"Verified"** y verás en los logs de TuReparto:

```
📥 Verificación recibida: mode=subscribe, token=tu_token_seguro_aqui
✅ Verificación exitosa! Meta ha confirmado el webhook.
```

### 3.4 ❌ Si la verificación falla

Revisa estos puntos:

| Problema | Solución |
|----------|----------|
| **"Token de verificación inválido"** | El `VERIFY_TOKEN` en TuReparto no coincide con el que pusiste en Meta |
| **"Callback URL no válida"** | Cloudflare Tunnel no está corriendo o la URL está mal escrita |
| **"Timeout"** | Cloudflare Tunnel no alcanza tu servidor — verifica que `./tureparto` esté corriendo |
| **Error genérico** | Revisa los logs de TuReparto y cloudflared para más detalles |

---

## 🧭 Paso 4: Suscribirse a eventos

Después de verificar, configura qué eventos quieres recibir:

1. En la misma sección **"Webhook"** del panel de Meta
2. En **"Webhook Fields"**, haz clic en **"Manage"**
3. Selecciona los campos que necesitas:

   | Campo | Descripción | ¿Necesario? |
   |-------|-------------|-------------|
   | ✅ **messages** | Mensajes entrantes de texto, imágenes, etc. | **Sí** |
   | ✅ **message_deliveries** | Confirmación de que un mensaje se entregó | Recomendado |
   | ✅ **message_reads** | Confirmación de lectura | Recomendado |
   | ✅ **message_echoes** | Mensajes enviados DESDE tu negocio | Opcional |
   | ✅ **message_reactions** | Reacciones a mensajes | Opcional |

4. Haz clic en **"Subscribe"**

---

## 🧭 Paso 5: Obtener credenciales

### Phone Number ID

1. Ve a **"Getting Started"** en el panel de WhatsApp
2. Verás un número de teléfono temporal asignado (o puedes agregar uno propio)
3. Anota el **"Phone Number ID"** (lo necesitarás para enviar mensajes)

### Token de acceso permanente

El token temporal expira cada 24 horas. Para uso en producción:

1. En **"Getting Started"**, usa el token temporal para pruebas
2. Para producción, genera un **Token Permanente**:
   - Ve a **"App Settings"** → **"Advanced"** → **"App Secret"**
   - Usa el **App ID** y **App Secret** para generar un token de larga duración via [Graph API](https://developers.facebook.com/docs/facebook-login/access-tokens/refreshing)

---

## 🧭 Paso 6: Enviar mensaje de prueba

### Desde el panel de Meta

1. En **WhatsApp → Getting Started**
2. En la sección **"Send a test message"**
3. Selecciona el número de teléfono de prueba
4. Escribe un mensaje y envía
5. En los logs de TuReparto deberías ver:

```
📩 Webhook recibido:
{
  "Entry": [...]
}
💬 De: 5215512345678 | Mensaje: Hola, es una prueba
```

### Desde tu celular

1. Agrega el número de WhatsApp Business a tus contactos
2. Envía un mensaje desde WhatsApp
3. El mensaje debería aparecer en los logs del servidor

---

## 🧭 Paso 7: Enviar respuestas (próximamente)

Una vez que recibes mensajes, puedes responder usando la API de Meta:

```bash
curl -X POST "https://graph.facebook.com/v21.0/PHONE_NUMBER_ID/messages" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "messaging_product": "whatsapp",
    "to": "5215512345678",
    "type": "text",
    "text": { "body": "¡Hola! Gracias por contactarnos." }
  }'
```

> Próximamente implementaremos respuestas automáticas en TuReparto.

---

## 📊 Diagrama de flujo completo

```
TU CELULAR                META                   CLOUDFLARE            TUREPARTO
    │                       │                        │                    │
    │  Envías WhatsApp      │                        │                    │
    │──────────────────────►│                        │                    │
    │                       │                        │                    │
    │                       │  POST /webhook         │                    │
    │                       │───────────────────────►│                    │
    │                       │                        │  localhost:3000    │
    │                       │                        │───────────────────►│
    │                       │                        │                    │
    │                       │                        │              Log: 💬
    │                       │                        │              De: 52...
    │                       │                        │              Msg: Hola
    │                       │                        │                    │
    │                       │                  200 OK│                    │
    │                       │◄───────────────────────│                    │
    │                       │◄───────────────────────│────────────────────│
    │                       │                        │                    │
```

---

## 🔗 Enlaces útiles

| Recurso | URL |
|---------|-----|
| Meta for Developers | https://developers.facebook.com/ |
| WhatsApp Cloud API Docs | https://developers.facebook.com/docs/whatsapp/cloud-api |
| WhatsApp API Reference | https://developers.facebook.com/docs/whatsapp/cloud-api/reference |
| Graph API Explorer | https://developers.facebook.com/tools/explorer/ |
| Cloudflare Tunnel Docs | https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/ |
