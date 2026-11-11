// Usage: node benchmarks/compare.cjs /absolute/path/to/config.json
// Config: {"corpus":"...", "out":"...new empty directory...",
// "variants":{"baseline":"...checkout...", "ours":"...checkout...",
// "asvedr":"...checkout...", "llehouerou":"...checkout..."}}
// Pin and review the checkouts first. The script never clones, edits or installs them.
const fs = require('fs');
const path = require('path');
const cp = require('child_process');
const crypto = require('crypto');
const config = JSON.parse(fs.readFileSync(process.argv[2], 'utf8'));
const out = path.resolve(config.out);
if (fs.existsSync(out)) throw Error('Output must be a new directory; preserve prior samples.');
fs.mkdirSync(out, {recursive: true});
const names = Object.keys(config.variants);
const files = ['pink90.mp3', 'classic.mp3', 'mpeg2.mp3'];
const hashes = Object.fromEntries(files.map(file => [file, crypto.createHash('sha256').update(fs.readFileSync(path.join(config.corpus, file))).digest('hex')]));
const run = (command, args, options = {}) => {
 const p = cp.spawnSync(command, args, {encoding: 'utf8', timeout: 180000, ...options});
 if (p.status !== 0) throw Error(`${command}: ${p.error || ''}\n${p.stdout}\n${p.stderr}`);
 return p.stdout;
};
const metadata = {go: run('go', ['version']).trim(), platform: process.platform, arch: process.arch, inputs: hashes, variants: {}};
for (const name of names) {
 if (!/^[a-z0-9-]+$/.test(name)) throw Error('Use simple variant names.');
 const checkout = path.resolve(config.variants[name]);
 const mod = fs.readFileSync(path.join(checkout, 'go.mod'), 'utf8').match(/^module\s+(\S+)/m)[1];
 const wd = path.join(out, name); fs.mkdirSync(wd);
 const make = name === 'llehouerou' ? 'mp3.NewDecoderWithOptions(r,mp3.DecoderOptions{Gapless:false})' : 'mp3.NewDecoder(r)';
 fs.writeFileSync(path.join(wd, 'go.mod'), `module study\n\ngo 1.20\nrequire ${mod} v0.0.0\nreplace ${mod} => ${JSON.stringify(checkout.replaceAll('\\','/'))}\n`);
 fs.writeFileSync(path.join(wd, 'study_test.go'), `package study
import("bytes";"io";"os";"testing";mp3 "${mod}")
type reader struct{io.Reader}
func BenchmarkDecode(b *testing.B){data,e:=os.ReadFile(os.Getenv("MP3_FILE"));if e!=nil{b.Fatal(e)};buf:=make([]byte,32768);b.ReportAllocs();b.SetBytes(int64(len(data)));b.ResetTimer();for i:=0;i<b.N;i++{r:=reader{bytes.NewReader(data)};d,e:=${make};if e!=nil{b.Fatal(e)};n,e:=io.CopyBuffer(io.Discard,d,buf);if e!=nil||n==0{b.Fatalf("decode %d %v",n,e)}}}
`);
 run('go', ['test', '-mod=mod', '-c', '-o', path.join(wd, 'study.exe')], {cwd: wd, env: {...process.env, GOWORK: 'off'}});
 const rev = cp.spawnSync('git', ['-C', checkout, 'rev-parse', 'HEAD'], {encoding:'utf8'});
 const dirty = cp.spawnSync('git', ['-C', checkout, 'status', '--porcelain'], {encoding:'utf8'});
 metadata.variants[name] = {checkout, module: mod, revision: rev.stdout.trim(), status: dirty.stdout};
}
fs.writeFileSync(path.join(out, 'metadata.json'), JSON.stringify(metadata,null,2));
const rows = [];
for (const file of files) for (let round=0; round<(file==='pink90.mp3'?5:3); round++) {
 const offset=round%names.length;
 for (const name of names.slice(offset).concat(names.slice(0,offset))) {
  const log=run(path.join(out,name,'study.exe'),['-test.run','^$','-test.bench','^BenchmarkDecode$','-test.benchtime',file==='pink90.mp3'?'5x':'3x'],{env:{...process.env,MP3_FILE:path.resolve(config.corpus,file)}});
  fs.writeFileSync(path.join(out,`${file}-${round+1}-${name}.log`),log);
  const m=log.match(/BenchmarkDecode-\d+\s+\d+\s+(\d+) ns\/op\s+[\d.]+ MB\/s\s+(\d+) B\/op\s+(\d+) allocs\/op/);
  if(!m)throw Error('Cannot parse benchmark: '+log);
  rows.push({file,round:round+1,name,ns:+m[1],bytes:+m[2],allocs:+m[3]});
  fs.writeFileSync(path.join(out,'results.json'),JSON.stringify(rows,null,2));
  console.log(rows.at(-1));
 }
}
