import { forwardRef } from "react";
import { cn } from "../../utils/cn";

export const Input = forwardRef<
	HTMLInputElement,
	React.ComponentProps<"input">
>(({ className, type, ...props }, ref) => {
	return (
		<input
			type={type}
			className={cn(
				`flex h-10 w-full rounded-md border border-border-default border-solid bg-transparent px-3
				text-base shadow-sm transition-colors
				file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-content-primary
				placeholder:text-content-secondary
				focus-visible:outline-hidden focus-visible:ring-2 focus-visible:ring-content-link
				disabled:cursor-not-allowed disabled:opacity-50 md:text-sm text-inherit`,
				className,
			)}
			ref={ref}
			{...props}
		/>
	);
});
