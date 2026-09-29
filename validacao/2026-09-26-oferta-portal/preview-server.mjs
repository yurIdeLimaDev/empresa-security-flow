// Local-only UI validation. Synthetic authentication is NEVER part of deploy.
import { createServer } from "node:http";
import { readFile } from "node:fs/promises";
import worker from "../../landing-page/dist/cloudflare/vexkeep-landing-worker.mjs";
const sample = JSON.parse(await readFile(new URL("../../landing-page/examples/customer-case.example.json", import.meta.url)));
const { userId: _userId, revision: _revision, approvalSha256: _approval, ...publicCase } = sample;
void _userId; void _revision; void _approval;
createServer(async (req, res) => {
  try {
    const url = new URL(req.url, "http://127.0.0.1:4318");
    if (req.method !== "GET") { res.writeHead(405).end(); return; }
    let response;
    if (url.pathname === "/api/auth/get-session") response = Response.json({ user: { name: "Cliente fictício", email: "demo@example.invalid" } });
    else if (url.pathname === "/api/account/overview") response = Response.json({ cases: [publicCase], hasMore: false });
    else if (url.pathname.startsWith("/api/")) response = Response.json({ error: "Prévia sem serviços externos" }, { status: 503 });
    else response = await worker.fetch(new Request(url), {}, { waitUntil() {} });
    const headers = Object.fromEntries(response.headers);
    let body = Buffer.from(await response.arrayBuffer());
    if (headers["content-type"]?.includes("text/html")) {
      body = Buffer.from(body.toString().replace(/<script src="https:\/\/challenges\.cloudflare\.com[^<]*<\/script>/g, "").replace("</body>", '<div style="position:fixed;bottom:0;left:0;right:0;z-index:9999;background:#fff4ca;color:#543b08;padding:7px;text-align:center;font-size:12px">PRÉVIA LOCAL • DADOS FICTÍCIOS • SEM COBRANÇA</div></body>'));
    }
    res.writeHead(response.status, { ...headers, "cache-control": "no-store" }).end(body);
  } catch { res.writeHead(500).end("Local preview unavailable"); }
}).listen(4318, "127.0.0.1", () => console.log("Synthetic preview on http://127.0.0.1:4318"));
