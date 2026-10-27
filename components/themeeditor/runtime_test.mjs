import {readFileSync} from 'node:fs';
import {runInNewContext} from 'node:vm';
import assert from 'node:assert/strict';
import test from 'node:test';
const fixture=JSON.parse(readFileSync(0,'utf8'));
const authored=readFileSync(new URL('../../assets/js/src/components/theme-tools.js',import.meta.url),'utf8');
const source=authored.slice(0,authored.indexOf('  const instances'))+'globalThis.themeTest={valid,parseColor,paletteColor,derive,css,frame};})();';
function setup(){const listeners=new Map(),styles=[],warning={textContent:''},parent={postMessage(){}},classes=new Set(),context={URL,URLSearchParams,AbortController,location:{origin:'http://localhost'},parent,matchMedia:()=>({matches:false}),window:{addEventListener(name,handler){listeners.set(name,handler)}},document:{querySelector:()=>warning,createElement:()=>({textContent:''}),head:{append:style=>styles.push(style)},fonts:{ready:Promise.resolve(),check:()=>true}}};runInNewContext(source,context);return {api:context.themeTest,context,listeners,styles,warning,classes};}
const keys=fixture.tokens.map(t=>t.key);
for(const preset of fixture.presets)test(`complete snapshot and derived edits: ${preset.key}`,()=>{const {api}=setup();assert.equal(api.valid(preset.definition,keys),true);assert.ok(!api.css(preset.definition,fixture.fonts).includes('var('));const edited=structuredClone(preset.definition);edited.tokens.primary=api.parseColor('#165d9b');api.derive(edited,fixture.tokens);for(const token of fixture.tokens.filter(t=>!t.editable)){const actual=edited.tokens[token.key],expected=preset.edited.tokens[token.key];assert.equal(actual.space,expected.space);assert.ok(Math.abs(actual.alpha-expected.alpha)<1e-7);for(let i=0;i<3;i++)assert.ok(Math.abs(actual.channels[i]-expected.channels[i])<1e-7,token.key);}});
test('wire validation rejects unknown fields, channel counts and non-finite colors',()=>{const {api}=setup();const d=structuredClone(fixture.presets[0].definition);d.rawCSS='body{}';assert.equal(api.valid(d,keys),false);delete d.rawCSS;d.tokens.primary.channels.push(0);assert.equal(api.valid(d,keys),false);assert.throws(()=>api.parseColor('url(javascript:alert(1))'));assert.throws(()=>api.parseColor('oklch(NaN 0 0)'));assert.throws(()=>api.parseColor('color(srgb 2 0 0)'));});
test('unsafe provider URLs never become exported CSS',()=>{const {api}=setup(),d=structuredClone(fixture.presets[0].definition);d.typography.title={source:'upload',id:'font:abc'};assert.throws(()=>api.css(d,[...fixture.fonts,{source:'upload',id:'font:abc',families:['font'],weight:'400',style:'normal',url:'javascript:alert(1)',format:'woff2'}]));});
test('frame requires matching origin, source, instance, version and monotonic sequence',async()=>{const {api,context,listeners,styles,classes}=setup(),root={dataset:{theme:'goshtoso',themeFrame:JSON.stringify({id:'preview',fonts:fixture.fonts,tokens:fixture.tokens})},classList:{toggle(name,value){value?classes.add(name):classes.delete(name)}}};await api.frame(root);const send=listeners.get('message'),message={type:'goshtoso:theme-state',version:1,instanceId:'preview',sequence:1,definition:fixture.presets[0].definition,previewMode:'light'};await send({origin:'http://evil.test',source:context.parent,data:message});await send({origin:'http://localhost',source:{},data:message});assert.equal(styles.length,0);await send({origin:'http://localhost',source:context.parent,data:message});assert.equal(styles.length,1);const first=styles[0].textContent;await send({origin:'http://localhost',source:context.parent,data:{...message,sequence:1,previewMode:'dark'}});assert.equal(classes.has('dark'),false);await send({origin:'http://localhost',source:context.parent,data:{...message,sequence:2,previewMode:'dark'}});assert.equal(classes.has('dark'),true);await send({origin:'http://localhost',source:context.parent,data:{...message,sequence:3,version:2,previewMode:'light'}});assert.equal(classes.has('dark'),true);assert.equal(styles[0].textContent,first);});

test('asset fonts use opaque aliases without interpreting embedded family metadata',()=>{const {api}=setup(),d=structuredClone(fixture.presets[0].definition);d.typography.title={source:'upload',id:'font:abc'};const text=api.css(d,[...fixture.fonts,{source:'upload',id:'font:abc',families:['<embedded metadata>'],weight:'400',style:'normal',url:'/fonts/abc',format:'woff2'}]);assert.ok(text.includes('goshtoso-font-font-abc'));assert.ok(!text.includes('embedded metadata'));});

test('frame loads uploaded fonts even when no example heading uses the title role',async()=>{
 const {api,context,listeners,warning}=setup(),calls=[];
 context.document.fonts.load=async value=>{calls.push(value);return [{}];};
 const font={source:'upload',id:'font:unused-title',families:['Title'],url:'/fonts/title',format:'woff2',weight:'400',style:'normal'};
 const d=structuredClone(fixture.presets[0].definition);d.typography.title={source:font.source,id:font.id};
 await api.frame({dataset:{theme:'goshtoso',themeFrame:JSON.stringify({id:'preview',fonts:[...fixture.fonts,font],tokens:fixture.tokens})},classList:{toggle(){}}});
 await listeners.get('message')({origin:'http://localhost',source:context.parent,data:{type:'goshtoso:theme-state',version:1,instanceId:'preview',sequence:1,definition:d,previewMode:'light'}});
 await new Promise(resolve=>setImmediate(resolve));
 assert.equal(calls.length,1);assert.ok(calls[0].includes('goshtoso-font-font-unused-title'));assert.equal(warning.textContent,'');
});

test('palette choices materialize literal colors and reset to the baseline',()=>{
 const {api}=setup(),baseline={space:'srgb',channels:[.1,.2,.3],alpha:1},style={getPropertyValue:key=>key==='--color-blue-700'?'oklch(0.488 0.243 264.376)':''};
 const blue=api.paletteColor('blue-700',baseline,style);assert.equal(blue.space,'oklch');assert.equal(blue.channels[0],.488);
 assert.equal(api.paletteColor('white',baseline,style).channels[0],1);
 assert.equal(api.paletteColor('black',baseline,style).channels[0],0);
 assert.equal(api.paletteColor('#123456',baseline,style).space,'srgb');
 const reset=api.paletteColor('',baseline,style);assert.deepEqual(JSON.parse(JSON.stringify(reset)),baseline);assert.notEqual(reset,baseline);
 assert.equal(api.paletteColor('var(--color-primary)',baseline,style),null);
 assert.equal(api.paletteColor('missing-500',baseline,style),null);
});
