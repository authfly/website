<?php

declare(strict_types=1);

namespace App\View;

use Latte\Engine;
use Latte\Loaders\FileLoader;
use PhpFasty\Core\View\TemplateRendererInterface;

final class LatteRenderer implements TemplateRendererInterface
{
    private Engine $latte;

    public function __construct(string $templatesDir, string $cacheDir)
    {
        $this->latte = new Engine();
        $this->latte->setLoader(new FileLoader($templatesDir));
        $this->latte->setTempDirectory($cacheDir);
    }

    public function render(string $template, array $params = []): string
    {
        return $this->latte->renderToString($template, $params);
    }
}
