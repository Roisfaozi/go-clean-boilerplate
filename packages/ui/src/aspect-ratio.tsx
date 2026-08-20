import * as React from "react";

const AspectRatio = React.forwardRef<
	HTMLDivElement,
	React.ComponentProps<"div"> & { ratio?: number }
>(({ ratio = 1, style, className, children, ...props }, ref) => (
	<div
		ref={ref}
		className={className}
		style={{
			position: "relative",
			width: "100%",
			paddingBottom: `${(1 / ratio) * 100}%`,
			...style,
		}}
		{...props}
	>
		<div
			style={{
				position: "absolute",
				inset: 0,
			}}
		>
			{children}
		</div>
	</div>
));
AspectRatio.displayName = "AspectRatio";

export { AspectRatio };
