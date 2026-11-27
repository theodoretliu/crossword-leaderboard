import { Link } from "@tanstack/react-router";
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
    <div className="flex flex-row items-center justify-between gap-4 px-4 py-4">
      <Link to="/" className="text-xl font-semibold tracking-tight md:text-3xl">
        <h1>Teddy's Mini Leaderboard</h1>
      </Link>

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="outline"
            className="flex h-8 w-8 shrink-0 items-center justify-center p-0 md:hidden"
          >
            <Menu className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>

        <DropdownMenuContent>
          <DropdownMenuItem asChild>
            <Link to="/leaderboard">Leaderboard</Link>
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <Link to="/previous_weeks">Previous Weeks</Link>
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <Link to="/settings">Settings</Link>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <div className="hidden flex-row items-center gap-4 md:flex">
        <Link to="/leaderboard">Leaderboard</Link>
        <Link to="/previous_weeks">Previous Weeks</Link>
        <Link to="/settings">Settings</Link>
      </div>
    </div>
  );
};
