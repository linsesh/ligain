// Type declarations for custom modules

// SVG files as React components
declare module '*.svg' {
  import React from 'react';
  const content: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  export default content;
}

// PNG/JPG files
declare module '*.png' {
  const content: number;
  export default content;
}

declare module '*.jpg' {
  const content: string;
  export default content;
}

declare module '*.jpeg' {
  const content: string;
  export default content;
}

// Other asset types
declare module '*.webp' {
  const content: string;
  export default content;
}

declare module '*.gif' {
  const content: string;
  export default content;
}
