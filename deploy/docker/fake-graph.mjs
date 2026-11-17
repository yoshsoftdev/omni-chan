// Fake Meta Graph API for E2E stacks (docker-compose.e2e.yml). Never used in production:
// the backend refuses META_GRAPH_BASE_URL when APP_ENV=production.
//
// Conventions (stateless):
//   - WABA id "w<pnid>" owns phone number "<pnid>".
//   - Access token "bad-token" is rejected like an expired token (401 OAuthException).
//   - Every POST .../messages is recorded; GET /_sent?to=<wa_id> lists them for assertions.
import http from "node:http";
import { randomUUID } from "node:crypto";

const sent = [];

function reply(res, status, body) {
  res.writeHead(status, { "content-type": "application/json" });
  res.end(JSON.stringify(body));
}

function readBody(req) {
  return new Promise((resolve) => {
    let data = "";
    req.on("data", (c) => (data += c));
    req.on("end", () => resolve(data));
  });
}

const server = http.createServer(async (req, res) => {
  const url = new URL(req.url, "http://fake-graph");
  const parts = url.pathname.split("/").filter(Boolean); // [version, id, edge?]

  if (url.pathname === "/healthz") return reply(res, 200, { ok: true });
  if (url.pathname === "/_sent") {
    const to = url.searchParams.get("to")?.replace(/^\+/, "");
    return reply(res, 200, { data: to ? sent.filter((m) => m.to === to) : sent });
  }

  if (req.headers.authorization === "Bearer bad-token") {
    return reply(res, 401, {
      error: { message: "Error validating access token: Session has expired", type: "OAuthException", code: 190 },
    });
  }

  const [, id, edge] = parts;
  const body = req.method === "GET" || req.method === "DELETE" ? "" : await readBody(req);

  if (req.method === "POST" && edge === "messages") {
    const msg = JSON.parse(body || "{}");
    const wamid = `wamid.FAKE${randomUUID().replace(/-/g, "")}`;
    // Meta accepts "to" with or without "+"; store it like a wa_id so tests can match webhook numbers.
    sent.push({ phone_number_id: id, to: String(msg.to ?? "").replace(/^\+/, ""), type: msg.type, text: msg.text?.body ?? null, wamid, payload: msg });
    return reply(res, 200, {
      messaging_product: "whatsapp",
      contacts: [{ input: msg.to, wa_id: msg.to }],
      messages: [{ id: wamid }],
    });
  }
  if (req.method === "POST" || req.method === "DELETE") return reply(res, 200, { success: true });

  if (id === "me") return reply(res, 200, { id: "fake-user", name: "E2E" });
  if (edge === "phone_numbers") {
    const pnid = id.startsWith("w") ? id.slice(1) : id;
    return reply(res, 200, {
      data: [{ id: pnid, display_phone_number: "628110000001", verified_name: "E2E", quality_rating: "GREEN" }],
    });
  }
  if (edge) return reply(res, 200, { data: [], paging: {} }); // message_templates and other list edges

  return reply(res, 200, {
    id,
    name: "E2E WABA",
    display_phone_number: "628110000001",
    verified_name: "E2E",
    quality_rating: "GREEN",
    status: "ACTIVE",
    timezone_id: "1",
    message_template_namespace: "e2e",
    business_verification_status: "verified",
    health_status: { can_send_message: "AVAILABLE", entities: [] },
  });
});

server.listen(8080, "0.0.0.0", () => console.log("fake graph listening on :8080"));
