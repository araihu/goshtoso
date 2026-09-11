# Pair A1 browser evidence

Executed 2026-09-10 against the report worktree at `707126f3`. These are minimal synthetic fixtures using the actual authored factories and vendored runtime, not modified components or a full demo E2E run. Adjust the checkout, Playwright package and browser executable paths to reproduce on another machine. Run each fenced program as a `.cjs` file with Node.

## Watch return and proposed Select morph

```js
const {chromium}=require('/home/gui/.cache/ms-playwright-go/1.62.1/package');
const root='/tmp/gs-htmx4-alpine-inputs';
(async()=>{
 const browser=await chromium.launch({headless:true,executablePath:'/home/gui/.cache/ms-playwright/chromium_headless_shell-1234/chrome-linux/headless_shell',args:['--no-sandbox']});
 const page=await browser.newPage(); const errors=[]; page.on('pageerror',e=>errors.push(e.message));
 const cfg=Buffer.from(JSON.stringify({options:[{value:'a',label:'Alpha'}],selectedValues:['a'],alpineModel:'parentValue'})).toString('base64');
 await page.setContent(`<main x-data="{parentValue:'a'}"><section id="choice" x-data="goshtosoSelect($el)" data-select-config="${cfg}"><span x-text="selectedOption.label"></span></section></main>`);
 for (const p of ['assets/js/runtime/htmx.org/4.0.0/htmx.min.js','assets/js/runtime/htmx.org/4.0.0/hx-alpine-compat.js','assets/js/src/components/data.js','assets/js/src/components/select.js','assets/js/runtime/alpinejs/3.17.2/alpine.min.js']) await page.addScriptTag({path:root+'/'+p});
 await page.waitForFunction(()=>document.querySelector('span').textContent==='Alpha');
 console.log('watchReturns',await page.evaluate(()=>Alpine.$data(document.querySelector('#choice')).modelUnwatches.map(x=>typeof x)));
 console.log('morphResult',await page.evaluate(async()=>{const old=document.querySelector('#choice');const config=JSON.parse(atob(old.dataset.selectConfig)); config.options[0].label='Beta';await htmx.swap({sourceElement:old,target:old,text:`<section id="choice" x-data="goshtosoSelect($el)" data-select-config="${btoa(JSON.stringify(config))}"><span x-text="selectedOption.label"></span></section>`,swap:'outerMorph'});await Alpine.nextTick();const el=document.querySelector('#choice');return {same:el===old,label:el.textContent,datasetLabel:JSON.parse(atob(el.dataset.selectConfig)).options[0].label};}));
 console.log('errors',errors);await browser.close();
})().catch(e=>{console.error(e);process.exit(1)});
```

Successful output:

```text
watchReturns [ 'undefined', 'undefined' ]
morphResult { same: true, label: 'Alpha', datasetLabel: 'Beta' }
errors []
```

## Client Combobox fragment restoration

The fixture deliberately starts after document readiness, seeds sessionStorage, and inserts a client Combobox root through htmx. A synthetic pageshow is used as a positive control for the existing restoration function; it is not a BFCache simulation.

```js
const {chromium}=require('/home/gui/.cache/ms-playwright-go/1.62.1/package');
(async()=>{const browser=await chromium.launch({headless:true,executablePath:'/home/gui/.cache/ms-playwright/chromium_headless_shell-1234/chrome-linux/headless_shell',args:['--no-sandbox']});const page=await browser.newPage();await page.route('http://probe.local/**',r=>r.fulfill({contentType:'text/html',body:'<main id="mount"></main>'}));await page.goto('http://probe.local/');for(const p of ['assets/js/runtime/htmx.org/4.0.0/htmx.min.js','assets/js/runtime/htmx.org/4.0.0/hx-alpine-compat.js','assets/js/runtime/alpinejs/3.17.2/alpine.min.js','assets/js/src/components/combobox-client.js'])await page.addScriptTag({path:'/tmp/gs-htmx4-alpine-inputs/'+p});console.log(await page.evaluate(async()=>{sessionStorage.setItem('goshtoso:combobox:probe:selected','["a"]');const node=document.querySelector('#mount');await htmx.swap({sourceElement:node,target:node,swap:'innerHTML',text:'<div id="probe" data-combobox data-combobox-mode="client" data-combobox-name="pick" x-data="{isOpen:false}"><span data-combobox-trigger-label-outer>Pick</span><div data-combobox-body><li data-combobox-option data-value="a"><span data-combobox-option-label>Alpha</span></li></div></div>'});await Alpine.nextTick();const before=document.querySelectorAll('#probe input').length;window.dispatchEvent(new Event('pageshow'));return {hiddenAfterFragment:before,hiddenAfterPageShow:document.querySelector('#probe input')?.value};}));await browser.close()})().catch(e=>{console.error(e);process.exit(1)});
```

Output:

```text
{ hiddenAfterFragment: 0, hiddenAfterPageShow: 'a' }
```

## Scoped unit command

```sh
go test ./components/checkbox ./components/combobox ./components/fileinput ./components/form/... ./components/palette ./components/radio ./components/range ./components/rating ./components/schemaform ./components/schematree ./components/search ./components/select ./components/structuredinput ./components/tagslist ./components/textarea ./components/textinput ./components/toggle -count=1
```

All 18 packages passed. These tests validate the reviewed baseline; they do not establish that the proposed refactors work.
