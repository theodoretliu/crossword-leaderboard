import Link from "next/link";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Menu } from "lucide-react";

export const Header = () => {
  return (
    <div className="flex flex-row items-center justify-between px-4 py-4 gap-4">
      <Link
        href="/"
        className="text-xl font-semibold tracking-tight md:text-3xl"
      >
        <h1>Teddy's Mini Leaderboard</h1>
      </Link>

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="outline"
            className="flex items-center justify-center h-8 w-8 p-0 md:hidden shrink-0"
          >
            <Menu className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>

        <DropdownMenuContent>
          <DropdownMenuItem asChild>
            <Link href="/previous_weeks">Previous Weeks</Link>
          </DropdownMenuItem>

          <DropdownMenuItem asChild>
            <Link href="/settings">Settings</Link>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <div className="flex-row gap-4 items-center hidden md:flex">
        <Link href="/previous_weeks" passHref legacyBehavior>
          <a>Previous Weeks</a>
        </Link>

        <Link href="/settings" passHref legacyBehavior>
          <a>Settings</a>
        </Link>
      </div>
    </div>
  );
};
