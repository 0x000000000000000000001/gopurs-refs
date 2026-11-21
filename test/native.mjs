// Test the library through the real generated FFI bridge, including saved actions.
import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { copyFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const compiler = process.env.GOPURS_COMPILER_DIR || fileURLToPath(new URL('../../gopurs/', import.meta.url));
const load = name => import(pathToFileURL(join(compiler, 'output', name, 'index.js')));
const [{ prepareFfi }, { generateFfiBridge, ffiProblems }, { Tuple }, { Just, Nothing }, C] = await Promise.all([
    load('Gopurs.FfiSupport'), load('Gopurs.FfiBridge'), load('Data.Tuple'), load('Data.Maybe'),
    load('PureScript.Backend.Optimizer.CoreFn'),
]);
const a = new C.TypeVar('a'), b = new C.TypeVar('b');
const fn = (args, result) => new C.Func(args, result);
const effect = result => new C.ADT('Effect.Effect', ['Effect', 'Effect'], [result]);
const ref = new C.ADT('Effect.Ref.Ref', ['Effect', 'Ref', 'Ref'], [a]);
const stateValue = new C.Record(new C.Row([new Tuple('state', a), new Tuple('value', b)], Nothing.value));
const foreigns = [
    ['_new', fn([a], effect(ref))],
    ['newWithSelf', fn([fn([ref], a)], effect(ref))],
    ['read', fn([ref], effect(a))],
    ['modifyImpl', fn([fn([a], stateValue), ref], effect(b))],
    ['modify_', fn([fn([a], a), ref], effect(C.Unit.value))],
    ['write', fn([a, ref], effect(C.Unit.value))],
].map(([name, type]) => new Tuple(name, new Just(new C.ForAll(['a', 'b'], type))));
const source = readFileSync(new URL('../src/Effect/Ref.go', import.meta.url), 'utf8');
const prepared = prepareFfi({ moduleName: 'Effect.Ref', path: 'Effect/Ref.go' })('Effect_Ref_')(source)();
assert.deepEqual(ffiProblems('Effect_Ref')(prepared.decls)(foreigns), []);
const bridge = generateFfiBridge('Effect_Ref')([])(prepared.decls)(foreigns);
assert.match(bridge, /gopurs_runtime\.Func3/);
const workspace = mkdtempSync(join(tmpdir(), 'gopurs-effect-ref-'));
try {
    mkdirSync(join(workspace, 'gopurs_runtime'));
    copyFileSync(join(compiler, 'runtime/runtime.go'), join(workspace, 'gopurs_runtime/runtime.go'));
    copyFileSync(new URL('./native_test.go', import.meta.url), join(workspace, 'ref_test.go'));
    writeFileSync(join(workspace, 'go.mod'), 'module gopurs/output\n\ngo 1.22\n');
    writeFileSync(join(workspace, 'ref.go'), `package Effect_Ref\n${prepared.content}\n${bridge}\n`);
    execFileSync('go', ['test', '-count=1', '-race', ...process.argv.slice(2), '.'], { cwd: workspace, stdio: 'inherit' });
} finally {
    rmSync(workspace, { recursive: true, force: true });
}
