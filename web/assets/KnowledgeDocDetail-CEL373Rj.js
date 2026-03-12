import{n as Q,u as Z,r as a,j as e}from"./react-BkC6JIs-.js";import{b as ee,c as te,d as ne,p as I,e as oe,r as se}from"./knowledge-6baCAOqk.js";import{u as re}from"./index-DZNrraaf.js";import{$ as le,ao as ae,S as c,a$ as ie,ac as A,B as h,bm as ce,bp as de,b0 as pe,T as ge,aa as E,ae as he,a4 as z,am as xe,aG as me,aI as q,an as ue,at as fe,O,ai as ye,z as be,bc as je,bq as ke,br as we,a_ as ve,aD as $e,aF as Se}from"./antd-pd6as3Fe.js";import{M as Re,r as Ce,a as Ie}from"./markdown-FXq1K9bH.js";import"./client-bV8YVfSI.js";import"./charts-CceV33u2.js";const{Text:f,Title:ze}=ge;function P(d){return d.toLowerCase().replace(/[^\w\u4e00-\u9fff\s-]/g,"").replace(/\s+/g,"-").replace(/-+/g,"-").trim()}function Te(d){const r=[],g=d.split(`
`);let l=!1;const n=new Map;for(const k of g){if(k.trim().startsWith("```")){l=!l;continue}if(l)continue;const u=k.match(/^(#{1,6})\s+(.+)/);if(u){const v=u[1].length,w=u[2].replace(/[*_`~]/g,"").trim();let x=P(w);const y=n.get(x)??0;n.set(x,y+1),y>0&&(x=`${x}-${y}`),r.push({level:v,text:w,slug:x})}}return r}function Be(d,r){const g=new Blob([d],{type:"text/markdown;charset=utf-8"}),l=URL.createObjectURL(g),n=document.createElement("a");n.href=l,n.download=r,document.body.appendChild(n),n.click(),document.body.removeChild(n),URL.revokeObjectURL(l)}const Oe=()=>{const d=re(),{id:r}=Q(),g=Z(),{message:l}=le.useApp(),[n,k]=a.useState(null),[u,v]=a.useState([]),[w,x]=a.useState([]),[y,F]=a.useState(!0),[b,T]=a.useState(d),[H,B]=a.useState(""),M=a.useRef(null),$=a.useCallback(async()=>{if(r)try{const[t,o,i]=await Promise.all([ee(r),te(r),ne()]);k(t),v(o),x(i)}catch{l.error("文档不存在"),g("/knowledge")}finally{F(!1)}},[r]);a.useEffect(()=>{$()},[$]);const m=a.useMemo(()=>n?.content?Te(n.content):[],[n?.content]),N=a.useMemo(()=>m.length>0?Math.min(...m.map(t=>t.level)):1,[m]);a.useEffect(()=>{if(m.length===0||b)return;const t=new IntersectionObserver(i=>{for(const s of i)s.isIntersecting&&B(s.target.id)},{rootMargin:"-80px 0px -60% 0px",threshold:.1});return M.current?.querySelectorAll("h1[id], h2[id], h3[id], h4[id], h5[id], h6[id]")?.forEach(i=>t.observe(i)),()=>t.disconnect()},[m,b]);const U=t=>{const o=i=>{for(const s of i){if(s.id===t)return s.display_name;if(s.children?.length){const p=o(s.children);if(p)return p}}return""};return o(w)},V=async()=>{if(r)try{await oe(r),l.success("文档已归档"),g("/knowledge")}catch{l.error("归档失败")}},W=async t=>{if(r)try{await se(r,t),l.success(`已回滚到 v${t}`),$()}catch{l.error("回滚失败")}},G=()=>{if(!n)return;const t=`${n.title.replace(/[/\\?%*:|"<>]/g,"_")}.md`;Be(n.content,t),l.success("下载成功")},K=t=>{const o=document.getElementById(t);o&&(o.scrollIntoView({behavior:"smooth",block:"start"}),B(t))},_=t=>{if(!t)return"";const o=new Date(t);return`${o.getFullYear()}-${String(o.getMonth()+1).padStart(2,"0")}-${String(o.getDate()).padStart(2,"0")} ${String(o.getHours()).padStart(2,"0")}:${String(o.getMinutes()).padStart(2,"0")}`},X=a.useMemo(()=>{const t=new Map,o=i=>({children:s,...p})=>{const J=String(s??"").replace(/[*_`~]/g,"").trim();let j=P(J);const C=t.get(j)??0;return t.set(j,C+1),C>0&&(j=`${j}-${C}`),e.jsx(i,{id:j,...p,children:s})};return{h1:o("h1"),h2:o("h2"),h3:o("h3"),h4:o("h4"),h5:o("h5"),h6:o("h6")}},[n?.content]);if(y||!n)return e.jsx("div",{style:{display:"flex",justifyContent:"center",alignItems:"center",height:"60vh"},children:e.jsx(ae,{size:"large"})});const Y=I(n.tags),S=I(n.qa_anchors),R=I(n.related_docs),D=U(n.domain_id),L=m.length>0;return e.jsxs("div",{style:{padding:d?"12px 12px 80px":"24px 24px 80px"},children:[L&&!d&&e.jsx("div",{style:{position:"fixed",left:0,top:72,zIndex:100,display:"flex",alignItems:"flex-start",transition:"transform 0.3s ease",transform:b?"translateX(-100%)":"translateX(0)"},children:e.jsxs("div",{style:{width:240,maxHeight:"calc(100vh - 100px)",overflowY:"auto",background:"rgba(20, 20, 34, 0.95)",backdropFilter:"blur(12px)",border:"1px solid #252540",borderLeft:"none",borderRadius:"0 12px 12px 0",padding:"16px 0",boxShadow:"4px 0 20px rgba(0,0,0,0.3)"},children:[e.jsxs("div",{style:{display:"flex",alignItems:"center",justifyContent:"space-between",padding:"0 16px 12px",borderBottom:"1px solid #252540",marginBottom:8},children:[e.jsxs(c,{size:6,children:[e.jsx(ie,{style:{color:"#1677ff",fontSize:14}}),e.jsx("span",{style:{fontSize:13,fontWeight:600,color:"rgba(255,255,255,0.85)"},children:"目录"})]}),e.jsx(A,{title:"收起目录",children:e.jsx(h,{type:"text",size:"small",icon:e.jsx(ce,{style:{fontSize:12}}),onClick:()=>T(!0),style:{color:"#888"}})})]}),e.jsx("div",{style:{padding:"0 8px"},children:m.map((t,o)=>{const i=(t.level-N)*14,s=H===t.slug;return e.jsx("div",{onClick:()=>K(t.slug),style:{padding:"6px 10px",paddingLeft:10+i,fontSize:t.level<=2?12:11,fontWeight:t.level<=2?500:400,color:s?"#1677ff":"rgba(255,255,255,0.65)",background:s?"rgba(22,119,255,0.1)":"transparent",borderLeft:s?"2px solid #1677ff":"2px solid transparent",borderRadius:"0 6px 6px 0",cursor:"pointer",transition:"all 0.2s",overflow:"hidden",textOverflow:"ellipsis",whiteSpace:"nowrap"},onMouseEnter:p=>{s||(p.currentTarget.style.color="rgba(255,255,255,0.85)",p.currentTarget.style.background="rgba(255,255,255,0.05)")},onMouseLeave:p=>{s||(p.currentTarget.style.color="rgba(255,255,255,0.65)",p.currentTarget.style.background="transparent")},title:t.text,children:t.text},`${t.slug}-${o}`)})})]})}),L&&!d&&b&&e.jsx(A,{title:"展开目录",placement:"right",children:e.jsx(h,{type:"primary",shape:"circle",size:"small",icon:e.jsx(de,{style:{fontSize:14}}),onClick:()=>T(!1),style:{position:"fixed",left:8,top:80,zIndex:100,boxShadow:"2px 2px 12px rgba(0,0,0,0.3)"}})}),e.jsx(pe,{items:[{title:e.jsx("a",{onClick:()=>g("/knowledge"),children:"知识库"})},...D?[{title:D}]:[],{title:n.title}],style:{marginBottom:16}}),e.jsx(ze,{level:2,style:{marginBottom:8},children:n.title}),e.jsxs(c,{size:12,wrap:!0,style:{marginBottom:12},children:[e.jsx(E,{count:`v${n.current_version}`,style:{backgroundColor:"#1677ff"}}),e.jsxs(f,{type:"secondary",children:[e.jsx(he,{style:{marginRight:4}}),n.updated_by||n.created_by," · ",_(n.updated_at)]}),n.status==="draft"&&e.jsx(z,{color:"default",children:"草稿"}),n.status==="archived"&&e.jsx(z,{color:"red",children:"已归档"})]}),e.jsx("div",{style:{marginBottom:16},children:e.jsx(c,{size:4,wrap:!0,children:Y.map(t=>e.jsx(z,{color:"blue",style:{borderRadius:4},children:t},t))})}),e.jsxs(c,{style:{marginBottom:24},children:[e.jsx(h,{type:"primary",icon:e.jsx(xe,{}),onClick:()=>g(`/knowledge/doc/${n.id}/edit`),children:"编辑"}),e.jsx(h,{icon:e.jsx(me,{}),onClick:G,children:"下载 MD"}),e.jsx(q,{title:"确认归档此文档？",onConfirm:V,children:e.jsx(h,{danger:!0,icon:e.jsx(ue,{}),children:"归档"})}),e.jsx(h,{icon:e.jsx(fe,{}),onClick:()=>g("/knowledge"),children:"返回"})]}),e.jsx(O,{style:{borderRadius:12,marginBottom:24},bodyStyle:{padding:"24px 32px"},children:e.jsx("div",{className:"knowledge-md-content",ref:M,children:e.jsx(Re,{remarkPlugins:[Ie],rehypePlugins:[Ce],components:X,children:n.content})})}),e.jsx(ye,{defaultActiveKey:["related"],style:{background:"transparent",border:"none"},items:[...R.length>0?[{key:"related",label:e.jsxs(c,{children:[e.jsx(je,{})," 关联文档 (",R.length,")"]}),children:e.jsx(c,{wrap:!0,children:R.map((t,o)=>e.jsx(O,{size:"small",hoverable:!0,style:{borderRadius:8,width:200},children:e.jsxs(c,{children:[e.jsx(be,{style:{color:"#1677ff"}}),e.jsx(f,{ellipsis:!0,style:{maxWidth:150},children:t})]})},o))})}]:[],...S.length>0?[{key:"qa",label:e.jsxs(c,{children:[e.jsx(ke,{})," 检索锚点 (",S.length,")"]}),children:e.jsx("ul",{style:{margin:0,paddingLeft:20},children:S.map((t,o)=>e.jsx("li",{style:{marginBottom:6},children:e.jsx(f,{children:t})},o))})}]:[],{key:"versions",label:e.jsxs(c,{children:[e.jsx(Se,{})," 版本历史 (",u.length,")"]}),children:e.jsx(we,{items:u.map((t,o)=>({color:o===0?"#1677ff":"#303030",children:e.jsxs("div",{style:{paddingBottom:4},children:[e.jsxs(c,{size:8,children:[e.jsx(E,{count:`v${t.version}`,style:{backgroundColor:o===0?"#1677ff":"#595959"}}),e.jsx(f,{type:"secondary",children:_(t.created_at)}),e.jsx(f,{type:"secondary",children:t.changed_by})]}),e.jsx("div",{style:{marginTop:4},children:e.jsx(f,{children:t.change_summary||"无变更说明"})}),e.jsxs(c,{style:{marginTop:4},children:[e.jsx(h,{size:"small",type:"link",icon:e.jsx(ve,{}),onClick:()=>{},children:"查看"}),o>0&&e.jsx(q,{title:`确认回滚到 v${t.version}？`,onConfirm:()=>W(t.version),children:e.jsx(h,{size:"small",type:"link",icon:e.jsx($e,{}),children:"回滚"})})]})]})}))})}]}),e.jsx("style",{children:`
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
      `})]})};export{Oe as default};
