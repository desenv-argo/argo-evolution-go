import { copyFile, mkdir, readFile, readdir, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const managerRoot = join(dirname(fileURLToPath(import.meta.url)), "..");
const assetsDirectory = join(managerRoot, "dist", "assets");

await mkdir(assetsDirectory, { recursive: true });
await copyFile(join(managerRoot, "src", "argo-manager.js"), join(assetsDirectory, "argo-manager.js"));
await copyFile(join(managerRoot, "public", "argo-brand.png"), join(assetsDirectory, "argo-brand.png"));
await copyFile(join(managerRoot, "public", "argo-watermark.png"), join(assetsDirectory, "argo-watermark.png"));

const bundleName = (await readdir(assetsDirectory)).find(
  (name) => /^index-.*\.js$/.test(name) && name !== "argo-manager.js",
);
if (!bundleName) {
  throw new Error("Evolution Manager bundle was not found in manager/dist/assets");
}

const bundlePath = join(assetsDirectory, bundleName);
let bundle = await readFile(bundlePath, "utf8");

const replacements = [
  {
    name: "dashboard",
    marker: 'function zz(){return m.jsx("argo-dashboard",{})}',
    source:
      'function zz(){return m.jsxs("div",{className:"p-6",children:[m.jsx("h1",{className:"mb-4 text-2xl font-bold text-foreground",children:"Dashboard"}),m.jsx("p",{className:"text-muted-foreground",children:"Dashboard content will be implemented here..."})]})}',
    target: 'function zz(){return m.jsx("argo-dashboard",{})}',
  },
  {
    name: "conversations",
    marker: 'function pL(){return m.jsx("argo-conversations",{})}',
    source:
      'function pL(){return m.jsxs("div",{className:"p-6",children:[m.jsx("h1",{className:"mb-4 text-2xl font-bold text-gray-900",children:"Mensagens"}),m.jsx("p",{className:"text-gray-600",children:"Messages sending will be implemented here..."})]})}',
    target: 'function pL(){return m.jsx("argo-conversations",{})}',
  },
  {
    name: "manager navigation",
    marker: '{to:"/manager/integrations",label:"Integrações",icon:mA}',
    source: 'const mj=[{to:"/manager",label:"Dashboard",icon:HR},{to:"/manager/instances",label:"Instâncias",icon:mA}];',
    target:
      'const mj=[{to:"/manager",label:"Dashboard",icon:HR},{to:"/manager/instances",label:"Instâncias",icon:mA},{to:"/manager/integrations",label:"Integrações",icon:mA},{to:"/manager/messages",label:"Conversas",icon:HR}];',
  },
  {
    name: "integrations page",
    marker: 'function ArgoIntegrationsPage(){return m.jsx("argo-integrations",{})}',
    source: 'function pL(){return m.jsx("argo-conversations",{})}function gL()',
    target:
      'function pL(){return m.jsx("argo-conversations",{})}function ArgoIntegrationsPage(){return m.jsx("argo-integrations",{})}function gL()',
  },
  {
    name: "integrations route",
    marker: 'path:"integrations",element:m.jsx(ArgoIntegrationsPage,{})',
    source: 'm.jsx(jn,{path:"messages",element:m.jsx(pL,{})}),m.jsx(jn,{path:"events",element:m.jsx(gL,{})})',
    target:
      'm.jsx(jn,{path:"messages",element:m.jsx(pL,{})}),m.jsx(jn,{path:"integrations",element:m.jsx(ArgoIntegrationsPage,{})}),m.jsx(jn,{path:"events",element:m.jsx(gL,{})})',
  },
  {
    name: "message lifecycle navigation",
    marker: '{to:"/manager/lifecycle",label:"Ciclo de mensagens",icon:HR}',
    source: '{to:"/manager/integrations",label:"Integrações",icon:mA},{to:"/manager/messages",label:"Conversas",icon:HR}',
    target: '{to:"/manager/integrations",label:"Integrações",icon:mA},{to:"/manager/lifecycle",label:"Ciclo de mensagens",icon:HR},{to:"/manager/messages",label:"Conversas",icon:HR}',
  },
  {
    name: "message lifecycle page",
    marker: 'function ArgoLifecyclePage(){return m.jsx("argo-message-lifecycle",{})}',
    source: 'function ArgoIntegrationsPage(){return m.jsx("argo-integrations",{})}function gL()',
    target: 'function ArgoIntegrationsPage(){return m.jsx("argo-integrations",{})}function ArgoLifecyclePage(){return m.jsx("argo-message-lifecycle",{})}function gL()',
  },
  {
    name: "message lifecycle route",
    marker: 'path:"lifecycle",element:m.jsx(ArgoLifecyclePage,{})',
    source: 'm.jsx(jn,{path:"integrations",element:m.jsx(ArgoIntegrationsPage,{})}),m.jsx(jn,{path:"events",element:m.jsx(gL,{})})',
    target: 'm.jsx(jn,{path:"integrations",element:m.jsx(ArgoIntegrationsPage,{})}),m.jsx(jn,{path:"lifecycle",element:m.jsx(ArgoLifecyclePage,{})}),m.jsx(jn,{path:"events",element:m.jsx(gL,{})})',
  },
];

for (const replacement of replacements) {
  if (bundle.includes(replacement.marker)) {
    continue;
  }
  if (!bundle.includes(replacement.source)) {
    throw new Error(`Could not locate ${replacement.name} placeholder in ${bundleName}`);
  }
  bundle = bundle.replace(replacement.source, replacement.target);
}

await writeFile(bundlePath, bundle);
console.log(`Argo Manager assets built and ${bundleName} integration verified.`);
