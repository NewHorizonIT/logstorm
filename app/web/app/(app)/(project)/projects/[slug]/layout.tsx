import { ProjectSidebar } from "@/components/layout/project-sidebar";

export default async function ProjectSlugLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;

  return (
    <>
      <ProjectSidebar projectSlug={slug} />
      <main className="flex flex-1 flex-col overflow-auto">{children}</main>
    </>
  );
}
