import { Metadata } from "next";
import PresentClient from "./PresentClient";

// Public token pages shouldn't be indexed generally, but we'll use a standard layout
export const metadata: Metadata = {
  title: "Presentation",
  description: "View presentation",
  robots: {
    index: false,
    follow: false,
  },
};

export default function PublicStoryPresentPage({ params }: { params: { token: string } }) {
  return <PresentClient token={params.token} />;
}
