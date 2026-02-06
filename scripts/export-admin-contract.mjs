import fs from 'fs';
import path from 'path';

const legacyRoot = process.env.LEGACY_ROOT || '/Users/magei/Code/telegram-redpacket';
const legacyServerRoot = path.join(legacyRoot, 'server', 'src');
const routesDir = path.join(legacyServerRoot, 'routes');
const serverIndexPath = path.join(legacyServerRoot, 'index.js');
const outputRoot = path.resolve(process.cwd(), 'docs', 'legacy-contract');

const MODULE_GROUPS = ['Auth', 'Users', 'Roles', 'Audit', 'SystemConfigs', 'Other'];

function ensureDir(dirPath) {
  fs.mkdirSync(dirPath, { recursive: true });
}

function readText(filePath) {
  return fs.readFileSync(filePath, 'utf8');
}

function splitTopLevelArgs(argStr) {
  const args = [];
  let current = '';
  let depth = 0;
  let inString = false;
  let stringQuote = '';
  for (let i = 0; i < argStr.length; i++) {
    const ch = argStr[i];
    const prev = argStr[i - 1];
    if (inString) {
      current += ch;
      if (ch === stringQuote && prev !== '\\') {
        inString = false;
        stringQuote = '';
      }
      continue;
    }
    if (ch === '"' || ch === "'" || ch === '`') {
      inString = true;
      stringQuote = ch;
      current += ch;
      continue;
    }
    if (ch === '(' || ch === '{' || ch === '[') depth++;
    if (ch === ')' || ch === '}' || ch === ']') depth--;
    if (ch === ',' && depth === 0) {
      args.push(current.trim());
      current = '';
      continue;
    }
    current += ch;
  }
  if (current.trim()) {
    args.push(current.trim());
  }
  return args;
}

function stripQuotes(raw) {
  if (!raw) return raw;
  const trimmed = raw.trim();
  if ((trimmed.startsWith('"') && trimmed.endsWith('"')) ||
      (trimmed.startsWith("'") && trimmed.endsWith("'")) ||
      (trimmed.startsWith('`') && trimmed.endsWith('`'))) {
    return trimmed.slice(1, -1);
  }
  return trimmed;
}

function findMatching(text, startIdx, openChar, closeChar) {
  let depth = 0;
  let inString = false;
  let stringQuote = '';
  for (let i = startIdx; i < text.length; i++) {
    const ch = text[i];
    const prev = text[i - 1];
    if (inString) {
      if (ch === stringQuote && prev !== '\\') {
        inString = false;
        stringQuote = '';
      }
      continue;
    }
    if (ch === '"' || ch === "'" || ch === '`') {
      inString = true;
      stringQuote = ch;
      continue;
    }
    if (ch === openChar) depth++;
    if (ch === closeChar) {
      depth--;
      if (depth === 0) return i;
    }
  }
  return -1;
}

function extractCall(text, startIdx) {
  const openIdx = text.indexOf('(', startIdx);
  if (openIdx === -1) return null;
  const closeIdx = findMatching(text, openIdx, '(', ')');
  if (closeIdx === -1) return null;
  return {
    openIdx,
    closeIdx,
    argsText: text.slice(openIdx + 1, closeIdx),
  };
}

function extractObjectKeys(objText) {
  const keys = new Set();
  if (!objText || !objText.trim().startsWith('{')) return keys;
  const inner = objText.trim().slice(1, -1);
  let depth = 0;
  let inString = false;
  let stringQuote = '';
  let token = '';
  for (let i = 0; i < inner.length; i++) {
    const ch = inner[i];
    const prev = inner[i - 1];
    if (inString) {
      token += ch;
      if (ch === stringQuote && prev !== '\\') {
        inString = false;
        stringQuote = '';
      }
      continue;
    }
    if (ch === '"' || ch === "'" || ch === '`') {
      inString = true;
      stringQuote = ch;
      token += ch;
      continue;
    }
    if (ch === '{' || ch === '[' || ch === '(') depth++;
    if (ch === '}' || ch === ']' || ch === ')') depth--;
    if (depth === 0 && ch === ':') {
      const key = token.trim();
      const cleaned = stripQuotes(key.replace(/[$?]/g, ''));
      if (cleaned) keys.add(cleaned);
      token = '';
      continue;
    }
    if (depth === 0 && ch === ',') {
      token = '';
      continue;
    }
    token += ch;
  }
  return keys;
}

function extractReqFields(body) {
  const fields = { body: new Set(), query: new Set(), params: new Set() };
  if (!body) return fields;
  const patterns = [
    { key: 'body', regex: /req\.body\.([a-zA-Z0-9_]+)/g },
    { key: 'query', regex: /req\.query\.([a-zA-Z0-9_]+)/g },
    { key: 'params', regex: /req\.params\.([a-zA-Z0-9_]+)/g },
    { key: 'body', regex: /req\.body\[['"]([^'"]+)['"]\]/g },
    { key: 'query', regex: /req\.query\[['"]([^'"]+)['"]\]/g },
    { key: 'params', regex: /req\.params\[['"]([^'"]+)['"]\]/g },
  ];
  for (const { key, regex } of patterns) {
    let match;
    while ((match = regex.exec(body)) !== null) {
      fields[key].add(match[1]);
    }
  }
  const destructuringTargets = [
    { key: 'body', regex: /(?:const|let|var)\s*\{\s*([^}]+)\}\s*=\s*req\.body/g },
    { key: 'query', regex: /(?:const|let|var)\s*\{\s*([^}]+)\}\s*=\s*req\.query/g },
    { key: 'params', regex: /(?:const|let|var)\s*\{\s*([^}]+)\}\s*=\s*req\.params/g },
  ];
  for (const { key, regex } of destructuringTargets) {
    let match;
    while ((match = regex.exec(body)) !== null) {
      const items = match[1].split(',').map((part) => part.trim().split(':')[0].trim());
      items.filter(Boolean).forEach((item) => fields[key].add(item));
    }
  }
  return fields;
}

function extractResponseKeys(body) {
  const keys = new Set();
  if (!body) return keys;
  let idx = 0;
  while (true) {
    const matchIdx = body.indexOf('res.json', idx);
    if (matchIdx === -1) break;
    const call = extractCall(body, matchIdx);
    if (!call) break;
    const args = splitTopLevelArgs(call.argsText);
    const firstArg = args[0];
    if (firstArg && firstArg.trim().startsWith('{')) {
      const callKeys = extractObjectKeys(firstArg);
      callKeys.forEach((k) => keys.add(k));
    }
    idx = call.closeIdx + 1;
  }
  return keys;
}

function extractErrorCodes(body) {
  const errors = new Set();
  if (!body) return errors;
  const regex = /res\.status\(\s*\d{3}\s*\)\.json\(\s*{[^}]*?error\s*:\s*['"`]([^'"`]+)['"`][^}]*?}\s*\)/g;
  let match;
  while ((match = regex.exec(body)) !== null) {
    errors.add(match[1]);
  }
  return errors;
}

function parseRouterUses(fileText) {
  const uses = [];
  let idx = 0;
  while (true) {
    const start = fileText.indexOf('router.use', idx);
    if (start === -1) break;
    const call = extractCall(fileText, start);
    if (!call) break;
    const args = splitTopLevelArgs(call.argsText);
    if (args.length > 0) {
      uses.push(args.slice(0).join(', '));
    }
    idx = call.closeIdx + 1;
  }
  return uses;
}

function parseRoutes(filePath, basePath, baseMiddlewares) {
  const fileText = readText(filePath);
  const fileLines = fileText.split('\n');
  const fileUses = parseRouterUses(fileText).join(', ');
  const routes = [];
  const methods = ['get', 'post', 'put', 'patch', 'delete'];
  for (const method of methods) {
    let idx = 0;
    const token = `router.${method}`;
    while (true) {
      const start = fileText.indexOf(token, idx);
      if (start === -1) break;
      const call = extractCall(fileText, start);
      if (!call) break;
      const args = splitTopLevelArgs(call.argsText);
      const pathArg = args[0] ? stripQuotes(args[0]) : '';
      const routeMiddlewares = args.slice(1, -1).join(', ');
      const handlerArg = args[args.length - 1] || '';
      const lineNumber = fileText.slice(0, start).split('\n').length;
      let handlerBody = '';
      if (handlerArg.includes('=>') || handlerArg.includes('function')) {
        const braceIdx = handlerArg.indexOf('{');
        if (braceIdx !== -1) {
          const closeIdx = findMatching(handlerArg, braceIdx, '{', '}');
          if (closeIdx !== -1) {
            handlerBody = handlerArg.slice(braceIdx + 1, closeIdx);
          }
        }
      }
      routes.push({
        method: method.toUpperCase(),
        path: buildFullPath(basePath, pathArg),
        routePath: pathArg,
        lineNumber,
        middlewareText: [baseMiddlewares, fileUses, routeMiddlewares].filter(Boolean).join(', '),
        handlerBody,
        handlerArg,
        filePath,
        fileLines,
      });
      idx = call.closeIdx + 1;
    }
  }
  return routes;
}

function buildFullPath(basePath, routePath) {
  if (!routePath || routePath === '/') return basePath;
  const base = basePath.endsWith('/') ? basePath.slice(0, -1) : basePath;
  const route = routePath.startsWith('/') ? routePath : `/${routePath}`;
  return `${base}${route}`;
}

function parseAdminMounts() {
  const text = readText(serverIndexPath);
  const mounts = [];
  let idx = 0;
  while (true) {
    const start = text.indexOf('app.use', idx);
    if (start === -1) break;
    const call = extractCall(text, start);
    if (!call) break;
    const args = splitTopLevelArgs(call.argsText);
    const basePath = stripQuotes(args[0] || '');
    if (!basePath.startsWith('/admin')) {
      idx = call.closeIdx + 1;
      continue;
    }
    const requireArg = args.find((arg) => arg.includes("require('./routes/"));
    const fileMatch = requireArg && requireArg.match(/require\(['"]\.\/routes\/([^'"]+)['"]\)/);
    if (!fileMatch) {
      idx = call.closeIdx + 1;
      continue;
    }
    const routeFile = fileMatch[1];
    const middlewares = args.slice(1, args.indexOf(requireArg)).join(', ');
    mounts.push({
      basePath,
      routeFile,
      middlewares,
    });
    idx = call.closeIdx + 1;
  }
  return mounts;
}

function hasMiddleware(mwText, name) {
  return mwText && mwText.includes(name);
}

function extractPerms(mwText) {
  if (!mwText) return [];
  const perms = new Set();
  const regex = /requireAdminPermission\(\s*['"`]([^'"`]+)['"`]\s*\)/g;
  let match;
  while ((match = regex.exec(mwText)) !== null) {
    perms.add(match[1]);
  }
  return Array.from(perms);
}

function inferGroup(endpoint) {
  const file = path.basename(endpoint.filePath, '.js');
  const fullPath = endpoint.path || '';
  if (file.toLowerCase().includes('auth') || fullPath.startsWith('/admin/auth')) return 'Auth';
  if (file.toLowerCase().includes('roles') || fullPath.includes('/admin-roles')) return 'Roles';
  if (file.toLowerCase().includes('audit') || fullPath.includes('/admin/audit')) return 'Audit';
  if (file.toLowerCase().includes('system') || fullPath.includes('/admin/system')) return 'SystemConfigs';
  if (file.toLowerCase().includes('users') || fullPath.includes('/admin/users') || fullPath.includes('/admin/admin-users')) {
    return 'Users';
  }
  return 'Other';
}

function formatRequest(fields) {
  const sections = [];
  if (fields.body.size > 0) {
    sections.push({
      name: 'body',
      keys: Array.from(fields.body),
    });
  }
  if (fields.query.size > 0) {
    sections.push({
      name: 'query',
      keys: Array.from(fields.query),
    });
  }
  if (fields.params.size > 0) {
    sections.push({
      name: 'params',
      keys: Array.from(fields.params),
    });
  }
  if (sections.length === 0) return 'TODO';
  const jsonLines = ['{'];
  sections.forEach((section, idx) => {
    const inner = section.keys.map((key) => `    "${key}": "TODO"`).join(',\n');
    jsonLines.push(`  "${section.name}": {`);
    if (inner) jsonLines.push(inner);
    jsonLines.push('  }' + (idx === sections.length - 1 ? '' : ','));
  });
  jsonLines.push('}');
  return jsonLines.join('\n');
}

function formatResponse(keys) {
  if (!keys || keys.size === 0) return 'TODO';
  const lines = ['{'];
  const entries = Array.from(keys).map((key) => `  "${key}": "TODO"`);
  lines.push(entries.join(',\n'));
  lines.push('}');
  return lines.join('\n');
}

function writeMarkdownFiles(endpoints, permissionsByGroup, responseNotes) {
  const files = {
    Auth: 'admin-auth.md',
    Users: 'admin-admin-users.md',
    Roles: 'admin-roles.md',
    Audit: 'admin-audit.md',
    SystemConfigs: 'admin-system-configs.md',
    Other: 'admin-other.md',
  };
  const todoByFile = {};
  for (const group of MODULE_GROUPS) {
    const fileName = files[group];
    const filePath = path.join(outputRoot, fileName);
    const groupEndpoints = endpoints.filter((ep) => ep.group === group);
    const lines = [`# ${group} Admin Contracts`, ''];
    const todoList = [];
    if (groupEndpoints.length === 0) {
      lines.push('暂无接口。', '');
    }
    for (const ep of groupEndpoints) {
      lines.push(`## ${ep.method} ${ep.path}`);
      lines.push('');
      lines.push(`- Auth: ${ep.auth}`);
      lines.push(`- Permission: ${ep.perm}`);
      lines.push(`- RateLimit: ${ep.rateLimit}`);
      lines.push(`- Source: ${ep.source}`);
      lines.push('');
      lines.push('### Request(JSON)');
      lines.push('```json');
      lines.push(ep.requestTemplate);
      lines.push('```');
      lines.push('');
      lines.push('### Response(JSON)');
      lines.push('```json');
      lines.push(ep.responseTemplate);
      lines.push('```');
      lines.push('');
      lines.push('### Errors');
      if (ep.errors.length === 0) {
        lines.push('- TODO');
        todoList.push(`Errors: ${ep.method} ${ep.path}`);
      } else {
        ep.errors.forEach((err) => lines.push(`- ${err}`));
      }
      lines.push('');
      if (ep.requestTemplate === 'TODO') {
        todoList.push(`Request: ${ep.method} ${ep.path}`);
      }
      if (ep.responseTemplate === 'TODO') {
        todoList.push(`Response: ${ep.method} ${ep.path}`);
      }
      if (ep.perm === 'TODO') {
        todoList.push(`Permission: ${ep.method} ${ep.path}`);
      }
      if (ep.rateLimit === 'todo') {
        todoList.push(`RateLimit: ${ep.method} ${ep.path}`);
      }
    }
    lines.push('---');
    lines.push('');
    lines.push('## 待补齐清单');
    if (todoList.length === 0) {
      lines.push('- 无');
    } else {
      todoList.forEach((item) => lines.push(`- ${item}`));
    }
    lines.push('');
    fs.writeFileSync(filePath, lines.join('\n'), 'utf8');
    todoByFile[fileName] = todoList;
  }

  const responsePath = path.join(outputRoot, 'response-format.md');
  const responseLines = [
    '# Response Format (Legacy Admin)',
    '',
    '## 现状',
    responseNotes.length ? responseNotes.map((note) => `- ${note}`).join('\n') : '- 未发现统一响应格式（多处直接 res.json(result) 或返回 { error, message }）',
    '',
    '## TODO',
    '- 明确是否需要统一结构（全局拦截或统一包装）',
    '- 补充 traceId/traceID 约定',
    '',
    '## 建议格式',
    '```json',
    '{',
    '  "code": "OK",',
    '  "message": "success",',
    '  "data": {},',
    '  "traceId": "TODO"',
    '}',
    '```',
    '',
  ];
  fs.writeFileSync(responsePath, responseLines.join('\n'), 'utf8');
  todoByFile['response-format.md'] = ['明确统一响应格式', '补充 traceId 约定'];

  const permissionPath = path.join(outputRoot, 'permissions.md');
  const permLines = ['# Permissions (Legacy Admin)', ''];
  for (const group of MODULE_GROUPS) {
    permLines.push(`## ${group}`);
    const perms = permissionsByGroup[group] || [];
    if (perms.length === 0) {
      permLines.push('- TODO');
    } else {
      perms.forEach((perm) => permLines.push(`- ${perm}`));
    }
    permLines.push('');
  }
  fs.writeFileSync(permissionPath, permLines.join('\n'), 'utf8');
  return todoByFile;
}

function generateIndex(endpoints) {
  const lines = ['# Admin API Index', '', '| Module | Method | Path | Auth | Permission | RateLimit | Source |', '| --- | --- | --- | --- | --- | --- | --- |'];
  for (const group of MODULE_GROUPS) {
    const groupEndpoints = endpoints.filter((ep) => ep.group === group);
    for (const ep of groupEndpoints) {
      lines.push(`| ${group} | ${ep.method} | ${ep.path} | ${ep.auth} | ${ep.perm} | ${ep.rateLimit} | ${ep.source} |`);
    }
  }
  fs.writeFileSync(path.join(outputRoot, 'index.md'), lines.join('\n'), 'utf8');
}

function main() {
  ensureDir(outputRoot);
  const mounts = parseAdminMounts();
  const endpoints = [];
  const permissionsByGroup = {};
  const responseNotes = new Set();
  const byRouteFile = new Map();

  for (const mount of mounts) {
    const routeFilePath = path.join(routesDir, `${mount.routeFile}.js`);
    if (!fs.existsSync(routeFilePath)) continue;
    byRouteFile.set(routeFilePath, mount);
    const routes = parseRoutes(routeFilePath, mount.basePath, mount.middlewares);
    for (const route of routes) {
      const middlewares = route.middlewareText || '';
      const auth = hasMiddleware(middlewares, 'requireAdminAuth') ? 'required' : 'optional';
      const perms = extractPerms(middlewares);
      const perm = perms.length > 0 ? perms.join(', ') : 'TODO';
      const rateLimit = hasMiddleware(middlewares, 'rateLimitLogin') ? 'yes' : 'no';
      const fields = extractReqFields(route.handlerBody);
      const responseKeys = extractResponseKeys(route.handlerBody);
      const errors = Array.from(extractErrorCodes(route.handlerBody));
      if (route.handlerBody.includes('res.json')) {
        responseNotes.add('存在直接 res.json(result) 与 res.json({ ... }) 混用');
      }
      if (route.handlerBody.includes('error:')) {
        responseNotes.add('错误响应通常包含 { error, message } 字段');
      }
      const group = inferGroup(route);
      if (!permissionsByGroup[group]) permissionsByGroup[group] = [];
      perms.forEach((p) => {
        if (!permissionsByGroup[group].includes(p)) permissionsByGroup[group].push(p);
      });
      endpoints.push({
        method: route.method,
        path: route.path,
        source: `server/src/routes/${path.basename(route.filePath)}:${route.lineNumber}`,
        group,
        auth,
        perm,
        rateLimit,
        requestTemplate: formatRequest(fields),
        responseTemplate: formatResponse(responseKeys),
        errors,
        filePath: route.filePath,
      });
    }
  }

  const adminEndpoints = endpoints.filter((ep) => ep.path.startsWith('/admin'));
  const endpointsJsonPath = path.join(outputRoot, '_admin_endpoints.json');
  fs.writeFileSync(endpointsJsonPath, JSON.stringify(adminEndpoints.map((ep) => ({
    method: ep.method,
    path: ep.path,
    source: ep.source,
    group: ep.group,
  })), null, 2), 'utf8');

  generateIndex(adminEndpoints);
  const todoByFile = writeMarkdownFiles(adminEndpoints, permissionsByGroup, Array.from(responseNotes));

  const summaryPath = path.join(outputRoot, '_todo_summary.json');
  fs.writeFileSync(summaryPath, JSON.stringify(todoByFile, null, 2), 'utf8');
}

main();
