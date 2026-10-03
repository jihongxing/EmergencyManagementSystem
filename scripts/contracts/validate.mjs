import { readFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { execFileSync } from 'node:child_process';
import SwaggerParser from '@apidevtools/swagger-parser';
import Ajv2020 from 'ajv/dist/2020.js';

export const root = resolve(import.meta.dirname, '../../constras/platform');
const contractRoot = resolve(root, '..');
const backendRoot = resolve(root, '../../backend');
export const probes = [
  { file: 'live.openapi.json', path: '/health/live', responses: { 200: 'alive' } },
  { file: 'ready.openapi.json', path: '/health/ready', responses: { 200: 'ready', 503: 'not_ready' } },
];

export async function validateMachineSchemas(directory = contractRoot) {
  const ajv = new Ajv2020({ strict: true, allErrors: true });
  for (const [schemaName, files] of [
    ['behavior.schema.json', [
      '01-scope-and-identity.json', '02-access-and-evidence.json',
      '03-rules-and-records.json', '04-subscription.json', '05-platform.json',
    ]],
    ['identity.schema.json', ['identity/authorization.json']],
    ['bootstrap.schema.json', ['identity/bootstrap.json']],
    ['session.schema.json', ['identity/session.json']],
    ['members.schema.json', ['identity/members.json']],
    ['entry.schema.json', ['identity/entry.json']],
    ['identity-lifecycle.schema.json', ['identity/lifecycle.json']],
  ]) {
    const schema = JSON.parse(await readFile(resolve(import.meta.dirname, schemaName), 'utf8'));
    const validate = ajv.compile(schema);
    for (const file of files) {
      const contract = JSON.parse(await readFile(resolve(directory, file), 'utf8'));
      if (!validate(contract)) throw new Error(`${file}: ${ajv.errorsText(validate.errors)}`);
    }
  }
  const dataSchema = JSON.parse(await readFile(resolve(directory, 'identity/data.schema.json'), 'utf8'));
  const dataValidate = ajv.compile(dataSchema);
  const dataFixture = {
    version: 1,
    organization: {
      id: 'org_o1',
      kind: 'enterprise',
      name: 'Fixture Enterprise',
      externalKey: 'fixture-enterprise',
      status: 'active',
      createdAt: '2026-10-03T00:00:00Z',
    },
    member: {
      id: 'mem_m1',
      userId: 'usr_u1',
      organizationId: 'org_o1',
      loginId: 'admin@example.invalid',
      status: 'active',
      roles: ['enterprise_admin', 'enterprise_executor'],
      createdAt: '2026-10-03T00:00:00Z',
      updatedAt: '2026-10-03T00:00:00Z',
    },
    session: {
      id: 'ses_s1',
      memberId: 'mem_m1',
      client: 'web',
      createdAt: '2026-10-03T00:00:00Z',
      expiresAt: '2026-10-03T01:00:00Z',
      revokedAt: null,
    },
    passwordReset: {
      id: 'rst_r1',
      memberId: 'mem_m1',
      expiresAt: '2026-10-03T01:00:00Z',
      usedAt: null,
      revokedAt: null,
    },
  };
  if (!dataValidate(dataFixture)) throw new Error(`identity/data.schema.json: ${ajv.errorsText(dataValidate.errors)}`);
  await SwaggerParser.validate(resolve(directory, 'identity/identity.openapi.json'), { validate: { spec: true } });
  const identityDocument = JSON.parse(await readFile(resolve(directory, 'identity/identity.openapi.json'), 'utf8'));
  if (!identityDocument.paths?.['/v1/auth/login']?.post ||
      !identityDocument.paths?.['/v1/organizations/{organizationId}/members']?.post ||
      !identityDocument.paths?.['/v1/controlled/organizations/bootstrap']?.post) {
    throw new Error('identity.openapi.json: required P1-01 operations are missing');
  }
}

export async function validateContracts(directory = root, responseRecords) {
  const ajv = new Ajv2020({ strict: true, allErrors: true });
  const responseKeys = new Set();
  for (const probe of probes) {
    const location = resolve(directory, probe.file);
    const document = JSON.parse(await readFile(location, 'utf8'));
    await SwaggerParser.validate(location, { validate: { spec: true } });
    const operation = document.paths?.[probe.path]?.get;
    if (!operation) throw new Error(`${probe.file}: missing GET ${probe.path}`);
    for (const [code, expected] of Object.entries(probe.responses)) {
      const schema = operation.responses?.[code]?.content?.['application/json']?.schema;
      if (!schema) throw new Error(`${probe.file}: missing JSON response ${code}`);
      const valid = ajv.compile(schema);
      if (!valid({ status: expected })) throw new Error(`${probe.file}: ${code} rejects expected response: ${ajv.errorsText(valid.errors)}`);
      for (const body of [{ status: 'wrong' }, { status: expected, extra: true }, {}]) {
        if (valid(body)) throw new Error(`${probe.file}: ${code} accepts invalid response ${JSON.stringify(body)}`);
      }
      const key = `${probe.path}:${code}`;
      responseKeys.add(key);
      if (responseRecords !== undefined) {
        const record = responseRecords.find((item) => `${item.path}:${item.status}` === key);
        if (!record || responseRecords.filter((item) => `${item.path}:${item.status}` === key).length !== 1) {
          throw new Error(`Missing or duplicate real response for ${key}`);
        }
        if (!/^application\/json(?:\s*;|$)/i.test(record.contentType) || !valid(record.body)) {
          throw new Error(`Real response mismatch for ${key}: ${ajv.errorsText(valid.errors)}`);
        }
      }
    }
  }
  if (responseRecords !== undefined && (responseRecords.length !== responseKeys.size ||
      responseRecords.some((record) => !responseKeys.has(`${record.path}:${record.status}`)))) {
    throw new Error('Real handler returned an undocumented response');
  }
}

export function readRealResponses() {
  const output = execFileSync('pwsh', [
    '-NoProfile', '-Command',
    'go test ./internal/platform -run "^TestContractResponses$" -count=1 -v',
  ], { cwd: backendRoot, encoding: 'utf8', timeout: 120000 });
  const records = [...output.matchAll(/CONTRACT_RESPONSE (\{[^\r\n]+\})/g)]
    .map((match) => JSON.parse(match[1]));
  if (records.length !== 3) throw new Error(`Expected 3 real handler responses; got ${records.length}`);
  return records;
}
