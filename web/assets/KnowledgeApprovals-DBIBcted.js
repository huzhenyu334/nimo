import{u as W,r,j as e}from"./react-BkC6JIs-.js";import{l as G,g as H,m as Z,n as q}from"./knowledge-DIQcToVf.js";import{u as J}from"./index-BOiuucAq.js";import{a4 as O,at as Q,T as U,y as V,az as X,a7 as Y,B as b,ay as ee,L as B,S as u,a9 as z,o as A,W as v,Z as te,ac as se,ab as ne}from"./antd-hmFDXkJU.js";import{M as re,r as oe,a as ae}from"./markdown-XucLBvEK.js";import"./charts-BSomwbWx.js";const{Content:ie,Sider:de}=B,{Text:n,Title:I}=U,me=()=>{W();const{message:p}=O.useApp(),i=J(),[c,D]=r.useState([]),[M,$]=r.useState([]),[_,L]=r.useState(!0),[g,N]=r.useState("pending"),[o,x]=r.useState(""),[h,y]=r.useState(""),[w,f]=r.useState(!1),m=r.useCallback(async()=>{try{const[t,d]=await Promise.all([G(g==="all"?"":g),H()]);D(t),$(d),t.length>0&&!o&&x(t[0].id)}catch{}finally{L(!1)}},[g]);r.useEffect(()=>{m()},[m]);const s=c.find(t=>t.id===o),E=t=>{const d=T=>{for(const l of T){if(l.id===t)return l.display_name;if(l.children?.length){const a=d(l.children);if(a)return a}}return""};return d(M)},P=async()=>{if(o){f(!0);try{await q(o,h),p.success("审批通过！文档已入库"),y(""),m()}catch(t){p.error("审批失败: "+(t.response?.data?.message||t.message))}finally{f(!1)}}},F=async()=>{if(o){f(!0);try{await Z(o,h),p.warning("已驳回"),y(""),m()}catch(t){p.error("驳回失败: "+(t.response?.data?.message||t.message))}finally{f(!1)}}},k=t=>{if(!t)return"";const d=new Date(t),l=new Date().getTime()-d.getTime(),a=Math.floor(l/6e4);if(a<1)return"刚刚";if(a<60)return`${a}分钟前`;const j=Math.floor(a/60);return j<24?`${j}小时前`:`${Math.floor(j/24)}天前`},K={pending:"#faad14",approved:"#52c41a",rejected:"#ff4d4f"};if(_)return e.jsx("div",{style:{display:"flex",justifyContent:"center",alignItems:"center",height:"60vh"},children:e.jsx(Q,{size:"large"})});const R=c.filter(t=>t.status==="pending").length,S=t=>e.jsxs("div",{onClick:()=>x(t.id),style:{padding:"12px 16px",borderBottom:"1px solid #262626",cursor:"pointer",background:o===t.id?"#262626":"transparent",borderLeft:t.status==="pending"?"3px solid #1677ff":"3px solid transparent",transition:"all 0.15s ease"},children:[e.jsxs(u,{size:6,children:[e.jsx(z,{color:t.action==="create"?"blue":"green",style:{fontSize:11,margin:0},children:t.action==="create"?"新建":"更新"}),e.jsx(n,{strong:!0,ellipsis:!0,style:{maxWidth:i?200:180},children:t.title})]}),e.jsx("div",{style:{marginTop:4},children:e.jsxs(u,{size:8,children:[e.jsx(A,{style:{fontSize:12,color:"rgba(255,255,255,0.45)"}}),e.jsx(n,{type:"secondary",style:{fontSize:12},children:t.submitted_by}),e.jsx(n,{type:"secondary",style:{fontSize:12},children:k(t.created_at)})]})})]},t.id),C=()=>s?e.jsxs(e.Fragment,{children:[e.jsxs(v,{size:"small",style:{borderRadius:12,marginBottom:12},children:[e.jsx(I,{level:5,style:{margin:0,marginBottom:8},children:s.title}),e.jsxs(u,{size:i?8:12,wrap:!0,children:[e.jsxs(n,{type:"secondary",children:[e.jsx(A,{style:{marginRight:4}}),s.submitted_by]}),e.jsxs(n,{type:"secondary",children:["域: ",E(s.domain_id)]}),e.jsx(z,{color:K[s.status],children:s.status}),e.jsx(n,{type:"secondary",children:k(s.created_at)})]}),s.review_note&&e.jsxs("div",{style:{marginTop:8,padding:"8px 12px",background:"#262626",borderRadius:6},children:[e.jsx(n,{type:"secondary",style:{fontSize:12},children:"审批意见: "}),e.jsx(n,{children:s.review_note})]})]}),e.jsx(v,{style:{borderRadius:12,flex:1,overflow:"auto",marginBottom:12},styles:{body:{padding:i?"12px 14px":"20px 24px"}},children:e.jsx("div",{className:"knowledge-md-content",children:e.jsx(re,{remarkPlugins:[ae],rehypePlugins:[oe],children:s.content})})}),s.status==="pending"&&e.jsx(v,{size:"small",style:{borderRadius:12},children:e.jsxs(u,{style:{width:"100%"},direction:"vertical",size:8,children:[e.jsx(te.TextArea,{value:h,onChange:t=>y(t.target.value),placeholder:"审批意见 (可选)",rows:2,style:{borderRadius:8}}),e.jsxs("div",{style:{display:"flex",justifyContent:"flex-end",gap:8},children:[e.jsx(b,{danger:!0,icon:e.jsx(se,{}),onClick:F,loading:w,children:"驳回"}),e.jsx(b,{type:"primary",icon:e.jsx(ne,{}),onClick:P,loading:w,style:{background:"#52c41a",borderColor:"#52c41a"},children:"通过并入库"})]})]})})]}):e.jsx("div",{style:{display:"flex",justifyContent:"center",alignItems:"center",height:"100%"},children:e.jsx(n,{type:"secondary",children:i?"点击审批项查看详情":"选择左侧审批项查看详情"})});return e.jsxs("div",{style:{padding:i?12:24,height:"calc(100vh - 56px)",display:"flex",flexDirection:"column"},children:[e.jsx("div",{style:{marginBottom:16},children:e.jsxs(I,{level:4,style:{margin:0},children:[e.jsx(V,{style:{marginRight:8}}),"知识审批"]})}),e.jsx(X,{activeKey:g,onChange:t=>{N(t),x("")},items:[{key:"pending",label:`待审批${R>0?` (${R})`:""}`},{key:"approved",label:"已通过"},{key:"rejected",label:"已驳回"}],style:{marginBottom:16}}),c.length===0?e.jsx(Y,{description:e.jsx(n,{type:"secondary",children:"暂无审批记录"}),style:{marginTop:48}}):i?e.jsx(e.Fragment,{children:o?e.jsxs("div",{style:{flex:1,display:"flex",flexDirection:"column",overflow:"auto"},children:[e.jsx(b,{type:"text",icon:e.jsx(ee,{}),onClick:()=>x(""),style:{alignSelf:"flex-start",marginBottom:8},children:"返回列表"}),C()]}):e.jsx("div",{style:{flex:1,overflow:"auto",background:"#1f1f1f",borderRadius:12},children:c.map(S)})}):e.jsxs(B,{style:{flex:1,background:"transparent",overflow:"hidden"},children:[e.jsx(de,{width:320,style:{background:"#1f1f1f",borderRadius:12,overflow:"auto",marginRight:16},children:c.map(S)}),e.jsx(ie,{style:{overflow:"auto",display:"flex",flexDirection:"column"},children:C()})]}),e.jsx("style",{children:`
        .knowledge-md-content h1, .knowledge-md-content h2, .knowledge-md-content h3 {
          border-left: 3px solid #1677ff;
          padding-left: 12px;
          margin-top: 24px;
        }
        .knowledge-md-content table {
          width: 100%;
          border-collapse: collapse;
          margin: 16px 0;
        }
        .knowledge-md-content th, .knowledge-md-content td {
          border: 1px solid #303030;
          padding: 8px 12px;
        }
        .knowledge-md-content th {
          background: #262626;
        }
        .knowledge-md-content tr:nth-child(even) td {
          background: rgba(255,255,255,0.02);
        }
        .knowledge-md-content code {
          background: #262626;
          padding: 2px 6px;
          border-radius: 4px;
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
      `})]})};export{me as default};
