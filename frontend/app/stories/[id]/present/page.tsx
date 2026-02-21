import { PresentClient } from "./PresentClient";

export default async function PresentPage({ params }: { params: Promise<{ id: string }> }) {
  const resolvedParams = await params;
  return <PresentClient id={resolvedParams.id} />;
}
