// Tiny proxy to capture Claude Code's system prompt
// Usage: node capture-prompt.mjs
// Then run CC with: ANTHROPIC_BASE_URL=http://localhost:9999 claude ...

import http from "node:http";
import https from "node:https";
import fs from "node:fs";

const TARGET = "https://api.anthropic.com";
const PORT = 9999;
let captured = false;

const server = http.createServer((req, res) => {
  let body = [];
  req.on("data", (chunk) => body.push(chunk));
  req.on("end", () => {
    const bodyBuf = Buffer.concat(body);
    
    // Capture the first messages request (contains system prompt)
    if (req.url.includes("/messages") && !captured) {
      try {
        const parsed = JSON.parse(bodyBuf.toString());
        if (parsed.system) {
          const outPath = "/home/claw/.openclaw/workspace/tools/cc-system-prompt.json";
          fs.writeFileSync(outPath, JSON.stringify({
            system: parsed.system,
            model: parsed.model,
            tools: parsed.tools,
            max_tokens: parsed.max_tokens,
          }, null, 2));
          console.log(`✅ System prompt captured → ${outPath}`);
          console.log(`   System prompt length: ${JSON.stringify(parsed.system).length} chars`);
          console.log(`   Tools count: ${parsed.tools?.length || 0}`);
          captured = true;
        }
      } catch (e) { console.error("Parse error:", e.message); }
    }

    // Forward to real API
    const url = new URL(req.url, TARGET);
    const options = {
      hostname: url.hostname,
      port: 443,
      path: url.pathname + url.search,
      method: req.method,
      headers: { ...req.headers, host: url.hostname },
    };

    const proxy = https.request(options, (proxyRes) => {
      res.writeHead(proxyRes.statusCode, proxyRes.headers);
      proxyRes.pipe(res);
    });
    proxy.on("error", (e) => {
      console.error("Proxy error:", e.message);
      res.writeHead(502);
      res.end("Proxy error");
    });
    proxy.write(bodyBuf);
    proxy.end();
  });
});

server.listen(PORT, () => {
  console.log(`🔍 Proxy listening on http://localhost:${PORT}`);
  console.log(`   Run CC with: ANTHROPIC_BASE_URL=http://localhost:${PORT} claude "hello"`);
});
