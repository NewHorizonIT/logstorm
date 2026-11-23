import { DashboardContent } from "@/components/dashboard/dashboard-content";

export default async function DashboardPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  return <DashboardContent slug={slug} />;
}
