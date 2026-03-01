const http = require('http');
const { renderToHtml } = require('./ui');

function createResponse(res) {
  return {
    status(code) { res.statusCode = code; return this; },
    header(name, value) { res.setHeader(name, value); return this; },
    json(value) { this.header('Content-Type', 'application/json'); res.end(JSON.stringify(value)); },
    text(value) { this.header('Content-Type', 'text/plain; charset=utf-8'); res.end(String(value)); },
    html(value) { this.header('Content-Type', 'text/html; charset=utf-8'); res.end(typeof value === 'string' ? value : renderToHtml(value)); },
  };
}

function createServer() {
  const routes = [];
  const middlewares = [];
  const api = {
    use(fn) { middlewares.push(fn); },
    get(path, fn) { routes.push({ method: 'GET', path, fn }); },
    post(path, fn) { routes.push({ method: 'POST', path, fn }); },
    listen(port, host = '0.0.0.0') {
      return new Promise((resolve) => {
        const server = http.createServer(async (req, res) => {
          const response = createResponse(res);
          const request = { method: req.method, path: req.url.split('?')[0], headers: req.headers, raw: req };
          let idx = 0;
          const next = async () => {
            if (idx < middlewares.length) {
              const mw = middlewares[idx++];
              await mw(request, response, next);
            }
          };
          await next();
          if (res.writableEnded) return;
          const match = routes.find((r) => r.method === req.method && r.path === request.path);
          if (!match) {
            response.status(404).json({ error: 'Not Found', path: request.path });
            return;
          }
          try {
            await match.fn(request, response);
          } catch (err) {
            response.status(500).json({ error: err.message || 'Internal error' });
          }
        });
        server.listen(port, host, () => resolve({ close: () => server.close() }));
      });
    }
  };
  return api;
}

module.exports = { createServer };
