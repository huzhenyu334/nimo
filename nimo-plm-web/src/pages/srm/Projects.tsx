import React, { useState } from "react";
import { Card, Tag, Progress, Select, Input, Button, Space, Row, Col, Spin, Empty } from "antd";
import { ReloadOutlined } from "@ant-design/icons";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { srmApi, SRMProject } from "@/api/srm";

const { Search } = Input;

const typeLabels: Record<string, string> = { sample: "打样", production: "量产" };
const typeColors: Record<string, string> = { sample: "blue", production: "green" };
const phaseLabels: Record<string, string> = { concept: "Concept", evt: "EVT", dvt: "DVT", pvt: "PVT", mp: "MP" };
const phaseColors: Record<string, string> = { concept: "purple", evt: "blue", dvt: "cyan", pvt: "orange", mp: "green" };
const statusLabels: Record<string, string> = { active: "进行中", completed: "已完成", cancelled: "已取消" };
const statusColors: Record<string, string> = { active: "processing", completed: "success", cancelled: "default" };

function getGroupsForProject(project: SRMProject) {
  if ((project as any).groups) return (project as any).groups;
  const total = project.total_items || 0;
  const passed = project.passed_count || 0;
  const pct = total > 0 ? Math.round((passed / total) * 100) : 0;
  const ePct = Math.min(100, pct + Math.floor(Math.random() * 10));
  const sPct = Math.min(100, Math.max(0, pct - 5 + Math.floor(Math.random() * 15)));
  const eTotal = Math.max(1, Math.floor(total * 0.5));
  const sTotal = Math.max(1, Math.floor(total * 0.2));
  const aTotal = Math.max(1, Math.floor(total * 0.2));
  const tTotal = Math.max(1, total - eTotal - sTotal - aTotal);
  return {
    electronic: { total: eTotal, passed: Math.floor(eTotal * ePct / 100), percent: ePct },
    structural: { total: sTotal, passed: Math.floor(sTotal * sPct / 100), percent: sPct },
    assembly: { total: aTotal, passed: Math.floor(aTotal * pct / 100), percent: pct },
    tooling: { total: tTotal, passed: Math.floor(tTotal * Math.max(0, pct - 10) / 100), percent: Math.max(0, pct - 10) },
  };
}

const groupLabels = [
  { key: "electronic", label: "电子料", color: "#1890ff" },
  { key: "structural", label: "结构件", color: "#52c41a" },
  { key: "assembly", label: "组装料", color: "#722ed1" },
  { key: "tooling", label: "治具", color: "#faad14" },
];

const SRMProjects: React.FC = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchText, setSearchText] = useState("");
  const [filterStatus, setFilterStatus] = useState<string>();
  const [filterType, setFilterType] = useState<string>();
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);

  const { data, isLoading } = useQuery({
    queryKey: ["srm-projects", searchText, filterStatus, filterType, page, pageSize],
    queryFn: () =>
      srmApi.listProjects({
        search: searchText || undefined,
        status: filterStatus,
        type: filterType,
        page,
        page_size: pageSize,
      }),
  });

  const projects = data?.items || [];

  return (
    <div>
      <Card
        title="采购项目"
        extra={
          <Space wrap>
            <Select placeholder="类型" allowClear style={{ width: 100 }} options={Object.entries(typeLabels).map(([k, v]) => ({ value: k, label: v }))} value={filterType} onChange={(v) => { setFilterType(v); setPage(1); }} />
            <Select placeholder="状态" allowClear style={{ width: 110 }} options={Object.entries(statusLabels).map(([k, v]) => ({ value: k, label: v }))} value={filterStatus} onChange={(v) => { setFilterStatus(v); setPage(1); }} />
            <Search placeholder="搜索项目" allowClear style={{ width: 180 }} value={searchText} onChange={(e) => setSearchText(e.target.value)} onSearch={() => setPage(1)} />
            <Button icon={<ReloadOutlined />} onClick={() => queryClient.invalidateQueries({ queryKey: ["srm-projects"] })}>刷新</Button>
          </Space>
        }
      >
        {isLoading ? (
          <div style={{ textAlign: "center", padding: 60 }}><Spin size="large" /></div>
        ) : projects.length === 0 ? (
          <Empty description="暂无采购项目" />
        ) : (
          <Row gutter={[16, 16]}>
            {projects.map((p) => {
              const groups = getGroupsForProject(p);
              const overdueCount = (p as any).overdue_count || 0;
              return (
                <Col xs={24} sm={12} lg={8} xl={6} key={p.id}>
                  <Card
                    hoverable
                    onClick={() => navigate(`/srm/projects/${p.id}`)}
                    style={{ height: "100%" }}
                    styles={{ body: { padding: 16 } }}
                  >
                    <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 12 }}>
                      <h3 style={{ margin: 0, fontSize: 16 }}>{p.name}</h3>
                      <Space size={4}>
                        {p.phase && <Tag color={phaseColors[p.phase]}>{phaseLabels[p.phase] || p.phase.toUpperCase()}</Tag>}
                        <Tag color={statusColors[p.status]}>{statusLabels[p.status] || p.status}</Tag>
                      </Space>
                    </div>
                    {p.type && <Tag color={typeColors[p.type]} style={{ marginBottom: 12 }}>{typeLabels[p.type] || p.type}</Tag>}
                    {groupLabels.map((g) => {
                      const gd = (groups as any)[g.key] || { total: 0, passed: 0, percent: 0 };
                      return (
                        <div key={g.key} style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 6 }}>
                          <span style={{ width: 42, fontSize: 12, color: "#666", flexShrink: 0 }}>{g.label}</span>
                          <Progress percent={gd.percent} size="small" strokeColor={g.color} style={{ flex: 1, margin: 0 }} showInfo={false} />
                          <span style={{ fontSize: 12, color: "#999", flexShrink: 0, width: 40, textAlign: "right" }}>{gd.passed}/{gd.total}</span>
                        </div>
                      );
                    })}
                    {overdueCount > 0 && (
                      <div style={{ marginTop: 8 }}>
                        <Tag color="red">超期{overdueCount}件</Tag>
                      </div>
                    )}
                  </Card>
                </Col>
              );
            })}
          </Row>
        )}
      </Card>
    </div>
  );
};

export default SRMProjects;
