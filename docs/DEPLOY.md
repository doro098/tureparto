# 🚀 Guía de Despliegue — TuReparto

Guía completa para desplegar TuReparto en producción con Cloudflare Tunnel y buenas prácticas.

---

## 📋 Opciones de despliegue

| Opción | Descripción | ¿Recomendado? |
|--------|-------------|---------------|
| **Cloudflare Tunnel** | Túnel HTTPS a servidor local | ✅ Para producción |
| **Servidor VPS** | Desplegar directamente en un VPS | ✅ Alternativa sólida |
| **Docker** | Contenerizar la aplicación | 🔜 Próximamente |

---

## 🛡️ Opción 1: Cloudflare Tunnel (Recomendado)

### Requisitos

- [ ] cloudflared instalado
- [ ] Cuenta en Cloudflare
- [ ] Un dominio configurado en Cloudflare

### 1.1 Autenticar cloudflared

```bash
cloudflared tunnel login
```

Esto abre el navegador para autenticar con tu cuenta de Cloudflare.

### 1.2 Crear el túnel

```bash
cloudflared tunnel create tureparto
```

Esto genera:
- Un archivo de credenciales en `~/.cloudflared/<tunnel-id>.json`
- Un ID único para tu túnel

### 1.3 Configurar el túnel

Crea o edita `~/.cloudflared/config.yml`:

```yaml
tunnel: tureparto
credentials-file: /home/tu-usuario/.cloudflared/<tunnel-id>.json

ingress:
  - hostname: webhook.tudominio.com
    service: http://localhost:3000
  - service: http_status:404
```

### 1.4 Configurar DNS

```bash
# Asocia tu dominio al túnel
cloudflared tunnel route dns tureparto webhook.tudominio.com
```

### 1.5 Iniciar el túnel como servicio

```bash
# Instalar como servicio del sistema
sudo cloudflared service install

# Iniciar
sudo systemctl start cloudflared

# Verificar estado
sudo systemctl status cloudflared

# Habilitar en inicio
sudo systemctl enable cloudflared
```

### 1.6 Iniciar TuReparto como servicio

Crea `/etc/systemd/system/tureparto.service`:

```ini
[Unit]
Description=TuReparto WhatsApp Webhook
After=network.target

[Service]
Type=simple
User=juan
WorkingDirectory=/home/juan/tureparto
ExecStart=/home/juan/tureparto/tureparto
Restart=always
RestartSec=5
Environment=PORT=3000
Environment=VERIFY_TOKEN=tu_token_seguro_aqui

[Install]
WantedBy=multi-user.target
```

Activar el servicio:

```bash
sudo systemctl daemon-reload
sudo systemctl start tureparto
sudo systemctl enable tureparto
sudo systemctl status tureparto
```

### 1.7 Verificar el despliegue

```bash
# Verificar que el servicio está activo
sudo systemctl status tureparto

# Ver logs
sudo journalctl -u tureparto -f

# Probar endpoint
curl https://webhook.tudominio.com/
```

---

## 🖥️ Opción 2: VPS (Servidor Virtual)

### Requisitos

- VPS con Ubuntu/Debian
- Go 1.22+ instalado
- Nginx o Caddy como proxy reverso
- Certificado SSL (Let's Encrypt)

### 2.1 Compilar para servidor remoto

```bash
# En tu máquina local
cd /home/juan/tureparto
GOOS=linux GOARCH=amd64 go build -o tureparto-linux .

# O compilar directamente en el VPS
git clone <tu-repo> /opt/tureparto
cd /opt/tureparto
go build -o tureparto .
```

### 2.2 Configurar Nginx como proxy reverso

```nginx
server {
    listen 80;
    server_name webhook.tudominio.com;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 2.3 SSL con Certbot

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d webhook.tudominio.com
```

### 2.4 Servicio systemd (igual que arriba)

```bash
# Crear /etc/systemd/system/tureparto.service (ver arriba)
sudo systemctl daemon-reload
sudo systemctl start tureparto
sudo systemctl enable tureparto
```

---

## 🐳 Opción 3: Docker (Próximamente)

```dockerfile
# Dockerfile (próximamente)
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o tureparto .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/tureparto .
EXPOSE 3000
CMD ["./tureparto"]
```

---

## 🔒 Seguridad

### Tokens y variables sensibles

```bash
# NO hardcodees tokens en el código
# Usa variables de entorno siempre
export VERIFY_TOKEN="token_seguro_aleatorio"
export PORT="3000"
```

### Recomendaciones

| Práctica | Descripción |
|----------|-------------|
| 🔑 Token fuerte | Usa un token largo y aleatorio (> 20 caracteres) |
| 🔒 Firewall | Bloquea todo excepto puerto 22 (SSH) y 80/443 (web) |
| 📝 Logs | Revisa logs regularmente con `journalctl -u tureparto -f` |
| 🔄 Actualizaciones | Mantén Go y cloudflared actualizados |
| 👤 Usuario dedicado | No ejecutes como root, crea un usuario específico |

### Generar token seguro

```bash
# Generar token aleatorio de 32 caracteres
openssl rand -hex 32
# Ejemplo: a7f3c8e1b2d4f6a8c0e2f4a6b8c0d2e4f6a8c0e2f4a6b8c0d2e4f6a8c0e2f4
```

---

## 📊 Monitoreo

### Logs en tiempo real

```bash
# Logs de TuReparto
journalctl -u tureparto -f

# Logs de cloudflared
journalctl -u cloudflared -f
```

### Health check

Crea un monitor externo (ej: UptimeRobot, Pingdom) que verifique:
```
URL: https://webhook.tudominio.com/
Frecuencia: Cada 5 minutos
Tipo: HTTP
```

---

## 🔄 Actualización del servidor

```bash
# 1. Detener el servicio
sudo systemctl stop tureparto

# 2. Respaldar binario anterior
cp tureparto tureparto.backup

# 3. Compilar nueva versión
git pull
go build -o tureparto .

# 4. Iniciar servicio
sudo systemctl start tureparto

# 5. Verificar
sudo systemctl status tureparto
curl http://localhost:3000/
```

---

## 📋 Checklist de producción

- [ ] Servicio systemd configurado e iniciado
- [ ] Cloudflare Tunnel funcionando como servicio
- [ ] SSL/HTTPS configurado
- [ ] Variables de entorno configuradas (no hardcodeadas)
- [ ] Firewall configurado
- [ ] Monitoreo/configuración de logs
- [ ] Token de verificación cambiado del default
- [ ] Webhook verificado en Meta
- [ ] Suscripción a eventos de mensaje activa
- [ ] Prueba de mensaje exitosa

---

## 🐛 Solución de problemas comunes

### El servicio no arranca

```bash
# Ver errores
sudo journalctl -u tureparto -n 50

# Verificar puerto
sudo lsof -i :3000

# Probar manualmente
sudo -u juan /home/juan/tureparto/tureparto
```

### Cloudflare Tunnel no funciona

```bash
# Ver logs
sudo journalctl -u cloudflared -n 50

# Probar túnel manual
cloudflared tunnel run tureparto

# Verificar DNS
cloudflared tunnel route dns tureparto webhook.tudominio.com
```

### Meta no puede alcanzar el webhook

```bash
# 1. Verificar que cloudflared esté corriendo
ps aux | grep cloudflared

# 2. Probar la URL pública
curl https://webhook.tudominio.com/

# 3. Probar verificación
curl -s -G 'https://webhook.tudominio.com/webhook' \
  --data-urlencode 'hub.mode=subscribe' \
  --data-urlencode 'hub.verify_token=tu_token' \
  --data-urlencode 'hub.challenge=123'
```
