'use client';

import Link from 'next/link';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Plus, Presentation, MoreVertical } from 'lucide-react';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';

// Mock data for stories
const mockStories = [
  { id: '1', title: 'Q4 Performance Review', description: 'Analysis of Q4 sales and growth metrics.', slideCount: 5, lastModified: '2025-01-15' },
  { id: '2', title: 'Marketing Campaign Results', description: 'Impact of holiday season campaigns.', slideCount: 3, lastModified: '2025-01-20' },
];

export default function StoriesPage() {
  return (
    <div className="container mx-auto py-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">My Data Stories</h1>
          <p className="text-muted-foreground mt-2">
            Create AI-powered presentations from your dashboards.
          </p>
        </div>
        <Link href="/stories/new">
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            New Story
          </Button>
        </Link>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {/* Create New Card */}
        <Link href="/stories/new">
          <Card className="h-full border-dashed hover:border-primary transition-colors cursor-pointer flex flex-col items-center justify-center min-h-[200px] bg-muted/40">
            <div className="p-6 text-center">
              <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center mx-auto mb-4">
                <Plus className="h-6 w-6 text-primary" />
              </div>
              <h3 className="font-semibold text-lg">Create New Story</h3>
              <p className="text-sm text-muted-foreground mt-1">Start from scratch or use a template</p>
            </div>
          </Card>
        </Link>

        {/* Story Cards */}
        {mockStories.map((story) => (
          <Card key={story.id} className="group hover:shadow-md transition-shadow">
            <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-2">
              <div className="flex items-center gap-2">
                <div className="p-2 bg-primary/10 rounded-md">
                  <Presentation className="h-4 w-4 text-primary" />
                </div>
              </div>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon" className="h-8 w-8">
                    <MoreVertical className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem>Edit</DropdownMenuItem>
                  <DropdownMenuItem>Duplicate</DropdownMenuItem>
                  <DropdownMenuItem className="text-destructive">Delete</DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </CardHeader>
            <CardContent>
              <CardTitle className="text-xl mb-2 group-hover:text-primary transition-colors">
                <Link href={`/stories/${story.id}`}>{story.title}</Link>
              </CardTitle>
              <CardDescription className="line-clamp-2">
                {story.description}
              </CardDescription>
            </CardContent>
            <CardFooter className="flex justify-between text-sm text-muted-foreground border-t pt-4">
              <span>{story.slideCount} slides</span>
              <span>{story.lastModified}</span>
            </CardFooter>
          </Card>
        ))}
      </div>
    </div>
  );
}
