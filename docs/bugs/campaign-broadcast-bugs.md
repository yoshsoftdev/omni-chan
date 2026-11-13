1. template dengan gambar perlu image link untuk brodcast, 
walaupun sudah diberikan saat submit ke meta untuk di review
2. select template perlu view pagenation dengan fitur search
3. tidak bisa kirim jika template EN
4. Broadcast engine ngaco karena perubahan struktur contact

# Public invoke via webhook-service only (backend :8066 must stay private)
curl -X POST 'http://192.168.113.140:8060/api/v1/automation/endpoints/ept_2324a9c866ec4cc3b5e2e115e0cb10c9' \
  -H 'Authorization: Bearer <REDACTED_AUTOMATION_KEY>' \
  -H 'Content-Type: application/json' \
  -d '{"to":"+6282288812003","kode":"277353"}'

(base) ➜  ~ curl -X POST 'http://192.168.113.140:8060/api/v1/automation/endpoints/ept_8c8eae69dc3749cda5422094b2aa7266' \
  -H 'Authorization: Bearer <REDACTED_AUTOMATION_KEY>' \
  -H 'Content-Type: application/json' \
  -d '{"to":"+6282288812003","kode":"277353"}'
{"message":"button 0 URL mapping is required","success":false}%

curl -X POST 'http://192.168.113.140:8060/api/v1/automation/endpoints/ept_903b8755e42c4838b9ccd3ffc7c375c9' \
  -H 'Authorization: Bearer <REDACTED_AUTOMATION_KEY>' \
  -H 'Content-Type: application/json' \
  -d '{"to":"+6282288812003","text":"halo apa kabar ini pesan otomatis"}'