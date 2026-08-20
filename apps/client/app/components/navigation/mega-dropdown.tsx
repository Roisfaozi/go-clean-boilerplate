import { cn } from "@casbin/ui";
import { useState, useRef, useEffect } from "react";

interface MegaDropdownSection {
	title: string;
	items: {
		label: string;
		description?: string;
		href?: string;
		icon?: React.ReactNode;
		onClick?: () => void;
	}[];
}

interface MegaDropdownProps {
	trigger: React.ReactNode;
	sections: MegaDropdownSection[];
	footer?: React.ReactNode;
	className?: string;
}

export function MegaDropdown({
	trigger,
	sections,
	footer,
	className,
}: MegaDropdownProps) {
	const [open, setOpen] = useState(false);
	const ref = useRef<HTMLDivElement>(null);

	useEffect(() => {
		const handler = (e: MouseEvent) => {
			if (ref.current && !ref.current.contains(e.target as Node))
				setOpen(false);
		};
		document.addEventListener("mousedown", handler);
		return () => document.removeEventListener("mousedown", handler);
	}, []);

	return (
		<div ref={ref} className="relative">
			<div onClick={() => setOpen(!open)} className="cursor-pointer">
				{trigger}
			</div>
			{open && (
				<div
					className={cn(
						"bg-popover border-border p-card-pad animate-fade-in absolute top-full left-0 z-50 mt-2 min-w-[480px] rounded-lg border shadow-lg",
						className,
					)}
				>
					<div className="grid grid-cols-2 gap-6">
						{sections.map((section) => (
							<div key={section.title}>
								<p className="text-caption text-muted-foreground mb-2 font-semibold tracking-wider uppercase">
									{section.title}
								</p>
								<div className="space-y-1">
									{section.items.map((item) => (
										<a
											key={item.label}
											href={item.href || "#"}
											onClick={(e) => {
												if (item.onClick) {
													e.preventDefault();
													item.onClick();
												}
												setOpen(false);
											}}
											className="hover:bg-surface-hover flex items-start gap-3 rounded-md p-2 transition-colors"
										>
											{item.icon && (
												<span className="text-muted-foreground mt-0.5 shrink-0">
													{item.icon}
												</span>
											)}
											<div>
												<p className="text-body text-foreground font-medium">
													{item.label}
												</p>
												{item.description && (
													<p className="text-caption text-muted-foreground">
														{item.description}
													</p>
												)}
											</div>
										</a>
									))}
								</div>
							</div>
						))}
					</div>
					{footer && (
						<div className="border-border mt-4 border-t pt-4">{footer}</div>
					)}
				</div>
			)}
		</div>
	);
}
