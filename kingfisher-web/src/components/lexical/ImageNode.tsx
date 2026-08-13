import React, { useRef } from 'react';
import { DecoratorNode, type EditorConfig, type LexicalEditor, type LexicalNode, type NodeKey, type SerializedLexicalNode, type Spread } from 'lexical';
import { useLexicalComposerContext } from '@lexical/react/LexicalComposerContext';
import { docApi } from '../../api/doc';

/** 图片节点序列化结构：src + 可选 alt/width/height */
export type SerializedImageNode = Spread<
  { type: 'image'; src: string; altText?: string; width?: number; height?: number },
  SerializedLexicalNode
>;

/** 图片节点：渲染为 <img>，src 为空时显示上传占位（点击选择文件 → 调 /docs/upload） */
export class ImageNode extends DecoratorNode<React.ReactNode> {
  __src: string;
  __altText: string;
  __width?: number;
  __height?: number;

  static getType(): string {
    return 'image';
  }

  static clone(node: ImageNode): ImageNode {
    return new ImageNode(node.__src, node.__altText, node.__key, node.__width, node.__height);
  }

  constructor(src = '', altText = '', key?: NodeKey, width?: number, height?: number) {
    super(key);
    this.__src = src;
    this.__altText = altText;
    this.__width = width;
    this.__height = height;
  }

  getSrc(): string {
    return this.getLatest().__src;
  }
  getAltText(): string {
    return this.getLatest().__altText;
  }

  setSrc(src: string, altText = '') {
    const writable = this.getWritable();
    writable.__src = src;
    writable.__altText = altText;
  }

  isInline(): boolean {
    return false;
  }
  isKeyboardSelectable(): boolean {
    return true;
  }

  decorate(_editor: LexicalEditor, _config: EditorConfig): React.ReactNode {
    // 直接读私有字段（getSrc/getLatest 在 decorate 渲染期没有 active editor state 会抛错）
    return <ImageComponent node={this} src={this.__src} altText={this.__altText} />;
  }

  static importJSON(serialized: SerializedImageNode): ImageNode {
    return new ImageNode(serialized.src, serialized.altText || '', undefined, serialized.width, serialized.height).updateFromJSON(serialized) as ImageNode;
  }

  exportJSON(): SerializedImageNode {
    return {
      ...super.exportJSON(),
      type: 'image',
      src: this.getSrc(),
      altText: this.getAltText(),
      width: this.__width,
      height: this.__height,
    };
  }

  createDOM(): HTMLElement {
    return document.createElement('div');
  }
  updateDOM(): boolean {
    return false;
  }
}

export function $createImageNode(src = '', altText = ''): ImageNode {
  return new ImageNode(src, altText);
}

export function $isImageNode(node: LexicalNode | null | undefined): node is ImageNode {
  return node instanceof ImageNode;
}

/** 图片渲染：src 为空时点按选文件上传；否则显示 <img> */
function ImageComponent({ node, src, altText }: { node: ImageNode; src: string; altText: string }): React.ReactElement {
  const [editor] = useLexicalComposerContext();
  const fileRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = React.useState(false);
  const [error, setError] = React.useState('');

  const onPick = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) return;
    setUploading(true);
    setError('');
    docApi
      .upload(file)
      .then((r) => {
        const url = (r.data as { url?: string })?.url || '';
        if (!url) return;
        // 必须在 editor.update 内改节点（getWritable 需要 active update）
        editor.update(() => {
          node.setSrc(url, file.name);
        });
      })
      .catch(() => setError('上传失败'))
      .finally(() => setUploading(false));
  };

  if (!src) {
    return (
      <div className="lex-image-upload" onClick={() => fileRef.current?.click()}>
        <input ref={fileRef} type="file" accept="image/*" style={{ display: 'none' }} onChange={onPick} />
        {uploading ? <span>上传中…</span> : error ? <span className="lex-image-error">{error}</span> : <span>点击上传图片</span>}
      </div>
    );
  }
  return <img src={src} alt={altText} className="lex-image" />;
}
