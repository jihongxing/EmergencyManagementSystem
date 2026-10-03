import assert from 'node:assert/strict';
import { mkdir, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { test } from 'node:test';
import { probes, root, validateContracts, validateMachineSchemas, readRealResponses } from './validate.mjs';

test('valid contracts and rejected contract mutations', async () => {
  const realResponses = readRealResponses();
  await validateMachineSchemas();
  await validateContracts(root, realResponses);
  const fixture = await mkdtemp(join(tmpdir(), 'ems-openapi-'));
  try {
    const originals = {};
    for (const probe of probes) {
      originals[probe.file] = JSON.parse(await readFile(resolve(root, probe.file), 'utf8'));
      await writeFile(join(fixture, probe.file), JSON.stringify(originals[probe.file]));
    }
    const mutations = [
      ['dangling reference', (doc) => { doc.paths['/health/live'].get.responses['200'] = { $ref: '#/components/responses/Missing' }; }],
      ['missing required OpenAPI info', (doc) => { delete doc.info.version; }],
      ['unknown status field', (doc) => { doc.paths['/health/live'].get.responses['200'].content['application/json'].schema.additionalProperties = true; }],
      ['incorrect status schema', (doc) => { doc.paths['/health/live'].get.responses['200'].content['application/json'].schema.properties.status.const = 'wrong'; }],
    ];
    for (const [name, mutate] of mutations) {
      const document = structuredClone(originals['live.openapi.json']);
      mutate(document);
      await writeFile(join(fixture, 'live.openapi.json'), JSON.stringify(document));
      await assert.rejects(validateContracts(fixture), undefined, name);
    }
    await assert.rejects(
      validateContracts(root, realResponses.map((record) =>
        record.path === '/health/live' ? { ...record, body: { status: 'wrong' } } : record)),
      /Real response mismatch/,
    );
    await assert.rejects(
      validateContracts(root, realResponses.map((record) =>
        record.path === '/health/live' ? { ...record, contentType: 'text/plain' } : record)),
      /Real response mismatch/,
    );
    await assert.rejects(
      validateContracts(root, realResponses.map((record) =>
        record.path === '/health/live' ? { ...record, status: 201 } : record)),
      /Missing or duplicate real response/,
    );
  } finally {
    await rm(fixture, { recursive: true, force: true });
  }
});

test('behavior and identity JSON Schemas reject unknown and missing fields', async () => {
  const fixture = await mkdtemp(join(tmpdir(), 'ems-machine-schema-'));
  try {
    for (const file of ['01-scope-and-identity.json', '02-access-and-evidence.json',
      '03-rules-and-records.json', '04-subscription.json', '05-platform.json',
      'identity/authorization.json', 'identity/lifecycle.json', 'identity/bootstrap.json', 'identity/session.json', 'identity/members.json',
      'identity/data.schema.json', 'identity/identity.openapi.json', 'identity/entry.json']) {
      await mkdir(resolve(fixture, file, '..'), { recursive: true });
      await writeFile(resolve(fixture, file), await readFile(resolve(root, '..', file)));
    }
    const behavior = resolve(fixture, '01-scope-and-identity.json');
    const original = JSON.parse(await readFile(behavior, 'utf8'));
    await writeFile(behavior, JSON.stringify({ ...original, extra: true }));
    await assert.rejects(validateMachineSchemas(fixture), /additional properties/);
    await writeFile(behavior, JSON.stringify({ ...original, policies: [{ ...original.policies[0], facts: undefined }] }));
    await assert.rejects(validateMachineSchemas(fixture), /required property/);
    await writeFile(behavior, JSON.stringify(original));
    const identity = resolve(fixture, 'identity/authorization.json');
    const identityOriginal = JSON.parse(await readFile(identity, 'utf8'));
    identityOriginal.cases[0].member.extra = true;
    await writeFile(identity, JSON.stringify(identityOriginal));
    await assert.rejects(validateMachineSchemas(fixture), /additional properties/);
    await writeFile(identity, await readFile(resolve(root, '..', 'identity/authorization.json')));
    const lifecycle = resolve(fixture, 'identity/lifecycle.json');
    const lifecycleOriginal = JSON.parse(await readFile(lifecycle, 'utf8'));
    lifecycleOriginal.cases[0].expected.extra = true;
    await writeFile(lifecycle, JSON.stringify(lifecycleOriginal));
    await assert.rejects(validateMachineSchemas(fixture), /additional properties/);
  } finally {
    await rm(fixture, { recursive: true, force: true });
  }
});
