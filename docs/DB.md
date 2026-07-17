# 🗄️ Base de Datos — TuReparto

Documentación de la base de datos SQLite utilizada por TuReparto para persistir los mensajes de WhatsApp.

---

## 📍 Ubicación

Por defecto, la base de datos se crea como `tureparto.db` en el directorio del proyecto.

Podés cambiarlo con la variable de entorno `DB_PATH`:

```bash
export DB_PATH="/ruta/completa/mensajes.db"
./tureparto
```

---

## 📋 Esquema

```sql
CREATE TABLE messages (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    from_number  TEXT    NOT NULL,
    message_body TEXT    NOT NULL,
    received_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

| Columna | Tipo | Descripción |
|---------|------|-------------|
| `id` | `INTEGER` | Identificador único del mensaje (autoincremental) |
| `from_number` | `TEXT` | Número de teléfono del remitente (ej: `5215512345678`) |
| `message_body` | `TEXT` | Contenido del mensaje de texto |
| `received_at` | `DATETIME` | Fecha y hora de recepción (se setea automáticamente) |

---

## 🔍 Consultas útiles

### Desde la terminal con sqlite3

```bash
# Todos los mensajes
sqlite3 tureparto.db "SELECT * FROM messages;"

# Últimos 10 mensajes (más recientes primero)
sqlite3 tureparto.db "SELECT * FROM messages ORDER BY received_at DESC LIMIT 10;"

# Mensajes de un número específico
sqlite3 tureparto.db "SELECT * FROM messages WHERE from_number = '5215512345678';"

# Contar mensajes por remitente
sqlite3 tureparto.db "
  SELECT from_number, COUNT(*) as total
  FROM messages
  GROUP BY from_number
  ORDER BY total DESC;
"

# Mensajes de hoy
sqlite3 tureparto.db "
  SELECT * FROM messages
  WHERE date(received_at) = date('now');
"

# Mensajes no leídos (si agregás campo processed)
sqlite3 tureparto.db "SELECT * FROM messages;"
```

### Desde Python

```python
import sqlite3
import json
from datetime import datetime

# Conectar a la base de datos
conn = sqlite3.connect("tureparto.db")
conn.row_factory = sqlite3.Row

# Obtener todos los mensajes
cursor = conn.execute("SELECT * FROM messages ORDER BY received_at DESC")
mensajes = [dict(row) for row in cursor.fetchall()]

# Mostrar como JSON
print(json.dumps(mensajes, indent=2, ensure_ascii=False))

# Procesar cada mensaje
for msg in mensajes:
    print(f"📱 {msg['from_number']}")
    print(f"💬 {msg['message_body']}")
    print(f"⏰ {msg['received_at']}")
    print("---")

conn.close()
```

### Desde Node.js

```javascript
const sqlite3 = require('sqlite3').verbose();
const db = new sqlite3.Database('tureparto.db');

db.all("SELECT * FROM messages ORDER BY received_at DESC", [], (err, rows) => {
    if (err) throw err;
    console.log(JSON.stringify(rows, null, 2));
});

db.close();
```

---

## 🔄 Integración con tu script IA

El flujo típico con una IA:

```python
import sqlite3
import time

def procesar_mensajes():
    conn = sqlite3.connect("tureparto.db")
    conn.row_factory = sqlite3.Row

    # Traer mensajes que no fueron procesados aún
    # (asumiendo que agregaste una columna 'processed')
    cursor = conn.execute("""
        SELECT * FROM messages
        WHERE id NOT IN (SELECT message_id FROM processed)
        ORDER BY received_at ASC
    """)
    
    for row in cursor:
        mensaje = dict(row)
        # 👇 Acá va tu LLM para procesar el mensaje
        respuesta_ia = tu_llm(mensaje["message_body"])
        
        # Guardar que ya fue procesado
        conn.execute(
            "INSERT INTO processed (message_id, result) VALUES (?, ?)",
            (mensaje["id"], respuesta_ia)
        )
    
    conn.commit()
    conn.close()

# Loop: procesar cada 30 segundos
while True:
    procesar_mensajes()
    time.sleep(30)
```

---

## 📊 Extender el esquema

Podés agregar más columnas cuando lo necesites. Por ejemplo, para trackear procesamiento:

```bash
sqlite3 tureparto.db "ALTER TABLE messages ADD COLUMN processed INTEGER DEFAULT 0;"
sqlite3 tureparto.db "ALTER TABLE messages ADD COLUMN category TEXT;"
sqlite3 tureparto.db "ALTER TABLE messages ADD COLUMN tags TEXT;"
```

O crear tablas adicionales:

```sql
CREATE TABLE processed (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id INTEGER NOT NULL,
    result     TEXT,
    processed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(id)
);
```

---

## ⚠️ Tips

- **Backup**: Copiá el archivo `.db` para hacer backup: `cp tureparto.db backup.db`
- **Tamaño**: SQLite soporta hasta ~140TB, más que suficiente para mensajes de WhatsApp
- **Concurrencia**: SQLite soporta lecturas concurrentes. TuReparto escribe, vos podés leer al mismo tiempo
- **Portabilidad**: El archivo `.db` se puede abrir desde cualquier lenguaje (Python, Node, Go, etc.)
