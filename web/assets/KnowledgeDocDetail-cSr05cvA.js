import{n as Q,u as G,r,j as e}from"./react-BkC6JIs-.js";import{b as J,e as X,g as Z,p as S,h as ee,r as te}from"./knowledge-DIQcToVf.js";import{u as ne}from"./index-BOiuucAq.js";import{a4 as oe,at as se,b6 as re,T as le,S as c,aA as ie,af as z,aj as ae,a9 as $,B as x,ar as ce,aK as de,aM as L,as as ge,ay as he,W as D,an as pe,y as me,bf as ue,bt as fe,bu as xe,b4 as ye,aH as je,aJ as we}from"./antd-hmFDXkJU.js";import{M as ke,r as be,a as ve}from"./markdown-XucLBvEK.js";import"./charts-BSomwbWx.js";const{Text:m,Title:Re}=le;function A(a){return a.toLowerCase().replace(/[^\w\u4e00-\u9fff\s-]/g,"").replace(/\s+/g,"-").replace(/-+/g,"-").trim()}function Se(a){const l=[],d=a.split(`
`);let i=!1;const o=new Map;for(const j of d){if(j.trim().startsWith("```")){i=!i;continue}if(i)continue;const u=j.match(/^(#{1,6})\s+(.+)/);if(u){const k=u[1].length,w=u[2].replace(/[*_`~]/g,"").trim();let p=A(w);const y=o.get(p)??0;o.set(p,y+1),y>0&&(p=`${p}-${y}`),l.push({level:k,text:w,slug:p})}}return l}function $e(a,l){const d=new Blob([a],{type:"text/markdown;charset=utf-8"}),i=URL.createObjectURL(d),o=document.createElement("a");o.href=i,o.download=l,document.body.appendChild(o),o.click(),document.body.removeChild(o),URL.revokeObjectURL(i)}const ze=()=>{const a=ne(),{id:l}=Q(),d=G(),{message:i}=oe.useApp(),[o,j]=r.useState(null),[u,k]=r.useState([]),[w,p]=r.useState([]),[y,E]=r.useState(!0),[P,C]=r.useState(""),B=r.useRef(null),b=r.useCallback(async()=>{if(l)try{const[t,n,s]=await Promise.all([J(l),X(l),Z()]);j(t),k(n),p(s)}catch{i.error("文档不存在"),d("/knowledge")}finally{E(!1)}},[l]);r.useEffect(()=>{b()},[b]);const g=r.useMemo(()=>o?.content?Se(o.content):[],[o?.content]),H=r.useMemo(()=>g.length>0?Math.min(...g.map(t=>t.level)):1,[g]);r.useEffect(()=>{if(g.length===0||a)return;const t=()=>{const n=B.current?.querySelectorAll("h1[id], h2[id], h3[id], h4[id], h5[id], h6[id]");if(!n)return;let s="";for(const h of n)if(h.getBoundingClientRect().top<=100)s=h.id;else break;s&&C(s)};return window.addEventListener("scroll",t,{passive:!0}),t(),()=>window.removeEventListener("scroll",t)},[g,a]);const W=t=>{for(const n of w){if(n.id===t)return[{id:n.id,name:n.display_name}];if(n.children?.length){for(const s of n.children)if(s.id===t)return[{id:n.id,name:n.display_name},{id:s.id,name:s.display_name}]}}return[]},q=async()=>{if(l)try{await ee(l),i.success("文档已归档"),d("/knowledge")}catch{i.error("归档失败")}},U=async t=>{if(l)try{await te(l,t),i.success(`已回滚到 v${t}`),b()}catch{i.error("回滚失败")}},K=()=>{if(!o)return;const t=`${o.title.replace(/[/\\?%*:|"<>]/g,"_")}.md`;$e(o.content,t),i.success("下载成功")},O=t=>{const n=document.getElementById(t);if(n){const s=n.getBoundingClientRect().top+window.scrollY-80;window.scrollTo({top:s,behavior:"smooth"}),C(t)}},T=t=>{if(!t)return"";const n=new Date(t);return`${n.getFullYear()}-${String(n.getMonth()+1).padStart(2,"0")}-${String(n.getDate()).padStart(2,"0")} ${String(n.getHours()).padStart(2,"0")}:${String(n.getMinutes()).padStart(2,"0")}`},I=r.useRef([]);I.current=g.map(t=>t.slug);const _=r.useRef(0);_.current=0;const V=r.useMemo(()=>{const t=n=>({children:s,...h})=>{const f=_.current++,N=I.current[f]??A(String(s??"").replace(/[*_`~]/g,"").trim());return e.jsx(n,{id:N,...h,children:s})};return{h1:t("h1"),h2:t("h2"),h3:t("h3"),h4:t("h4"),h5:t("h5"),h6:t("h6")}},[g]);if(y||!o)return e.jsx("div",{style:{display:"flex",justifyContent:"center",alignItems:"center",height:"60vh"},children:e.jsx(se,{size:"large"})});const Y=S(o.tags),v=S(o.qa_anchors),R=S(o.related_docs),M=W(o.domain_id),F=g.length>0;return e.jsxs("div",{style:{padding:a?"12px 12px 80px":"24px 24px 80px"},children:[e.jsx(re,{items:[{title:e.jsx("a",{onClick:()=>d("/knowledge"),children:"知识库"})},{title:o.title}],style:{marginBottom:16}}),e.jsx(Re,{level:2,style:{marginBottom:4},children:o.title}),M.length>0&&e.jsx("div",{style:{marginBottom:8},children:e.jsxs(c,{size:4,children:[e.jsx(ie,{style:{color:"#1677ff",fontSize:13}}),M.map((t,n)=>e.jsxs("span",{children:[n>0&&e.jsx(m,{type:"secondary",style:{fontSize:13,margin:"0 2px"},children:"/"}),e.jsx("a",{style:{fontSize:13,color:"#1677ff"},onClick:()=>d(`/knowledge?domain=${t.id}&scope=${o.scope||"company"}`),children:t.name})]},t.id))]})}),e.jsxs(c,{size:12,wrap:!0,style:{marginBottom:12},children:[e.jsx(z,{count:`v${o.current_version}`,style:{backgroundColor:"#1677ff"}}),e.jsxs(m,{type:"secondary",children:[e.jsx(ae,{style:{marginRight:4}}),o.updated_by||o.created_by," · ",T(o.updated_at)]}),o.status==="draft"&&e.jsx($,{color:"default",children:"草稿"}),o.status==="archived"&&e.jsx($,{color:"red",children:"已归档"})]}),e.jsx("div",{style:{marginBottom:16},children:e.jsx(c,{size:4,wrap:!0,children:Y.map(t=>e.jsx($,{color:"blue",style:{borderRadius:4},children:t},t))})}),e.jsxs(c,{style:{marginBottom:24},children:[e.jsx(x,{type:"primary",icon:e.jsx(ce,{}),onClick:()=>d(`/knowledge/doc/${o.id}/edit`),children:"编辑"}),e.jsx(x,{icon:e.jsx(de,{}),onClick:K,children:"下载 MD"}),e.jsx(L,{title:"确认归档此文档？",onConfirm:q,children:e.jsx(x,{danger:!0,icon:e.jsx(ge,{}),children:"归档"})}),e.jsx(x,{icon:e.jsx(he,{}),onClick:()=>d("/knowledge"),children:"返回"})]}),e.jsxs("div",{style:{display:"flex",gap:24,alignItems:"flex-start",marginBottom:24},children:[F&&!a&&e.jsxs("div",{style:{position:"sticky",top:72,width:200,flexShrink:0,maxHeight:"calc(100vh - 96px)",overflowY:"auto"},children:[e.jsx("div",{style:{fontSize:12,fontWeight:600,color:"rgba(255,255,255,0.45)",textTransform:"uppercase",letterSpacing:1,marginBottom:8,paddingLeft:10},children:"目录"}),g.map((t,n)=>{const s=(t.level-H)*12,h=P===t.slug;return e.jsx("div",{onClick:()=>O(t.slug),style:{padding:"4px 10px",paddingLeft:10+s,fontSize:12,lineHeight:"20px",fontWeight:t.level<=2?500:400,color:h?"#1677ff":"rgba(255,255,255,0.5)",borderLeft:h?"2px solid #1677ff":"2px solid rgba(255,255,255,0.08)",cursor:"pointer",transition:"color 0.2s",overflow:"hidden",textOverflow:"ellipsis",whiteSpace:"nowrap"},onMouseEnter:f=>{h||(f.currentTarget.style.color="rgba(255,255,255,0.85)")},onMouseLeave:f=>{h||(f.currentTarget.style.color="rgba(255,255,255,0.5)")},title:t.text,children:t.text},`${t.slug}-${n}`)})]}),e.jsx(D,{style:{borderRadius:12,flex:1,minWidth:0},bodyStyle:{padding:a?"16px":"24px 32px"},children:e.jsx("div",{className:"knowledge-md-content",ref:B,children:e.jsx(ke,{remarkPlugins:[ve],rehypePlugins:[be],components:V,children:o.content})})})]}),e.jsx(pe,{defaultActiveKey:["related"],style:{background:"transparent",border:"none"},items:[...R.length>0?[{key:"related",label:e.jsxs(c,{children:[e.jsx(ue,{})," 关联文档 (",R.length,")"]}),children:e.jsx(c,{wrap:!0,children:R.map((t,n)=>e.jsx(D,{size:"small",hoverable:!0,style:{borderRadius:8,width:200},children:e.jsxs(c,{children:[e.jsx(me,{style:{color:"#1677ff"}}),e.jsx(m,{ellipsis:!0,style:{maxWidth:150},children:t})]})},n))})}]:[],...v.length>0?[{key:"qa",label:e.jsxs(c,{children:[e.jsx(fe,{})," 检索锚点 (",v.length,")"]}),children:e.jsx("ul",{style:{margin:0,paddingLeft:20},children:v.map((t,n)=>e.jsx("li",{style:{marginBottom:6},children:e.jsx(m,{children:t})},n))})}]:[],{key:"versions",label:e.jsxs(c,{children:[e.jsx(we,{})," 版本历史 (",u.length,")"]}),children:e.jsx(xe,{items:u.map((t,n)=>({color:n===0?"#1677ff":"#303030",children:e.jsxs("div",{style:{paddingBottom:4},children:[e.jsxs(c,{size:8,children:[e.jsx(z,{count:`v${t.version}`,style:{backgroundColor:n===0?"#1677ff":"#595959"}}),e.jsx(m,{type:"secondary",children:T(t.created_at)}),e.jsx(m,{type:"secondary",children:t.changed_by})]}),e.jsx("div",{style:{marginTop:4},children:e.jsx(m,{children:t.change_summary||"无变更说明"})}),e.jsxs(c,{style:{marginTop:4},children:[e.jsx(x,{size:"small",type:"link",icon:e.jsx(ye,{}),onClick:()=>{},children:"查看"}),n>0&&e.jsx(L,{title:`确认回滚到 v${t.version}？`,onConfirm:()=>U(t.version),children:e.jsx(x,{size:"small",type:"link",icon:e.jsx(je,{}),children:"回滚"})})]})]})}))})}]}),e.jsx("style",{children:`
        .knowledge-md-content h1, .knowledge-md-content h2, .knowledge-md-content h3 {
          border-left: 3px solid #1677ff;
          padding-left: 12px;
          margin-top: 24px;
        }
        .knowledge-md-content h1[id], .knowledge-md-content h2[id], .knowledge-md-content h3[id],
        .knowledge-md-content h4[id], .knowledge-md-content h5[id], .knowledge-md-content h6[id] {
          scroll-margin-top: 80px;
        }
        .knowledge-md-content table {
          width: 100%;
          border-collapse: collapse;
          margin: 16px 0;
        }
        .knowledge-md-content th, .knowledge-md-content td {
          border: 1px solid #303030;
          padding: 8px 12px;
          text-align: left;
        }
        .knowledge-md-content th {
          background: #262626;
          font-weight: 600;
        }
        .knowledge-md-content tr:nth-child(even) td {
          background: rgba(255,255,255,0.02);
        }
        .knowledge-md-content code {
          background: #262626;
          padding: 2px 6px;
          border-radius: 4px;
          font-size: 13px;
        }
        .knowledge-md-content pre {
          background: #262626;
          padding: 16px;
          border-radius: 8px;
          overflow-x: auto;
        }
        .knowledge-md-content pre code {
          background: transparent;
          padding: 0;
        }
        .knowledge-md-content blockquote {
          border-left: 3px solid #303030;
          padding-left: 16px;
          color: rgba(255,255,255,0.65);
          margin: 16px 0;
        }
        .knowledge-md-content ul, .knowledge-md-content ol {
          padding-left: 24px;
        }
        .knowledge-md-content li {
          margin-bottom: 4px;
        }
        .knowledge-md-content a {
          color: #1677ff;
        }
        .knowledge-md-content img {
          max-width: 100%;
          border-radius: 8px;
        }
      `})]})};export{ze as default};
