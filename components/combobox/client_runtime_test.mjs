import {readFileSync} from 'node:fs';
import {runInNewContext} from 'node:vm';
import assert from 'node:assert/strict';
import test from 'node:test';
class Element {
 constructor(tag='div',attributes={}){this.tagName=tag.toUpperCase();this.attributes=new Map(Object.entries(attributes));this.children=[];this.classList={add(){},remove(){}};this._text='';}
 setAttribute(name,value){this.attributes.set(name,String(value));}
 getAttribute(name){return this.attributes.get(name)??(typeof this[name]==='string'?this[name]:null);}
 hasAttribute(name){return this.attributes.has(name);}
 removeAttribute(name){this.attributes.delete(name);}
 get textContent(){return this._text+this.children.map(child=>child.textContent).join('');}
 set textContent(value){this._text=String(value);this.children=[];}
 append(...children){for(const child of children){child.parentElement=this;this.children.push(child);}}
 replaceChildren(...children){this.children=[];this.append(...children);}
 insertBefore(child,before){child.parentElement=this;const index=this.children.indexOf(before);this.children.splice(index<0?0:index,0,child);}
 get firstChild(){return this.children[0]||null;}
 remove(){this.parentElement.children=this.parentElement.children.filter(child=>child!==this);}
 matches(selector){const tag=selector.match(/^[a-z]+/);if(tag&&this.tagName!==tag[0].toUpperCase())return false;for(const match of selector.matchAll(/\[([\w-]+)(?:=(?:"([^"]*)"|([^\]]+)))?\]/g)){if(match[2]!==undefined||match[3]!==undefined){if(this.getAttribute(match[1])!==(match[2]??match[3]))return false;}else if(!this.hasAttribute(match[1]))return false;}return true;}
 closest(selector){for(let node=this;node;node=node.parentElement)if(node.matches(selector))return node;return null;}
 querySelectorAll(selector){return this.children.flatMap(child=>[...(child.matches(selector)?[child]:[]),...child.querySelectorAll(selector)]);}
 querySelector(selector){return this.querySelectorAll(selector)[0]||null;}
 dispatchEvent(){throw Error('controlled load emitted an unsolicited event');}
}
const code=readFileSync(new URL('../../assets/js/src/components/combobox-client.js',import.meta.url),'utf8');
function setup(){const root=new Element('div',{id:'font-picker','data-combobox':'','data-combobox-mode':'client','data-combobox-name':'font','data-combobox-persist':'true'}),body=new Element('div',{'data-combobox-body':''}),list=new Element('ul',{role:'listbox'}),search=new Element('input',{'data-combobox-search':''}),trigger=new Element('button',{id:'font-picker-trigger'}),label=new Element('span',{'data-combobox-trigger-label-outer':''});search.value='';const hidden=new Element('input',{type:'hidden',name:'font'});hidden.value='old';body.append(hidden,search,list);root.append(trigger,label,body);const listeners=new Map();const document={getElementById:id=>id==='font-picker'?root:id==='font-picker-trigger'?trigger:null,createElement:tag=>new Element(tag),addEventListener:(type,fn)=>listeners.set(type,fn)};const window={CSS:{escape:String},sessionStorage:{setItem(){throw Error('controlled load wrote a preference');}}};runInNewContext(code,{document,window,CustomEvent:class{},Set});return{root,list,search,label,load:detail=>listeners.get('combobox:load')({target:root,detail:{id:'font-picker',...detail}}),searchFor:text=>{search.value=text;listeners.get('input')({target:search});}};}
test('controlled options preserve labels, styles and selected values without events or storage',()=>{const s=setup();s.load({options:[{value:'roboto',label:'Roboto',labelAttrs:{style:'font-family:Roboto','data-font':'roboto',onclick:'unsafe'}},{value:'inter',label:'Inter'}],values:['roboto']});assert.equal(s.label.textContent,'Roboto');const label=s.list.children[0].querySelector('[data-combobox-option-label]');assert.equal(label.getAttribute('style'),'font-family:Roboto');assert.equal(label.getAttribute('data-font'),'roboto');assert.equal(label.getAttribute('onclick'),null);assert.deepEqual(s.root.querySelectorAll('input[type="hidden"][name="font"]').map(input=>input.value),['roboto']);s.searchFor('ROB');assert.equal(s.list.children[0].hidden,false);assert.equal(s.list.children[1].hidden,true);});
test('invalid replacement cannot partially replace the selected catalog',()=>{const s=setup();s.load({options:[{value:'a',label:'Font A'}],values:['a']});s.load({options:[{value:'b',label:'Font B'}],values:['unknown']});assert.equal(s.label.textContent,'Font A');assert.equal(s.list.children[0].getAttribute('data-value'),'a');s.load({values:['a','b']});assert.equal(s.label.textContent,'Font A');});
