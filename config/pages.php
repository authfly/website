<?php

declare(strict_types = 1);

return [
    '/' => [
        'template' => 'home.latte',
        'title' => '{landing.seo.title}',
        'description' => '{landing.seo.description}',
        'keywords' => '{landing.seo.keywords}',
        'data' => ['site', 'navigation', 'landing', 'blog', 'projects'],
    ],
    '/blog' => [
        'template' => 'blog.latte',
        'title' => '{blog.seo.title}',
        'description' => '{blog.seo.description}',
        'keywords' => '{blog.seo.keywords}',
        'data' => ['site', 'navigation', 'blog'],
    ],
    '/blog/' => [
        'template' => 'blog.latte',
        'title' => '{blog.seo.title}',
        'description' => '{blog.seo.description}',
        'keywords' => '{blog.seo.keywords}',
        'data' => ['site', 'navigation', 'blog'],
    ],
    '/blog/@slug' => [
        'template' => 'blog-post.latte',
        'title' => '{post.title} — {site.name}',
        'description' => '{post.excerpt}',
        'keywords' => '{post.seo_keywords}',
        'data' => ['site', 'navigation', 'blog'],
        'dynamic' => [
            'param' => 'slug',
            'dataset' => 'blog',
            'collection' => 'posts',
            'lookup' => 'slug',
            'item' => 'post',
        ],
    ],
    '/blog/@slug/' => [
        'template' => 'blog-post.latte',
        'title' => '{post.title} — {site.name}',
        'description' => '{post.excerpt}',
        'keywords' => '{post.seo_keywords}',
        'data' => ['site', 'navigation', 'blog'],
        'dynamic' => [
            'param' => 'slug',
            'dataset' => 'blog',
            'collection' => 'posts',
            'lookup' => 'slug',
            'item' => 'post',
        ],
    ],
    '/about' => [
        'template' => 'about.latte',
        'title' => '{about.seo.title}',
        'description' => '{about.seo.description}',
        'keywords' => '{about.seo.keywords}',
        'data' => ['site', 'navigation', 'about'],
    ],
    '/about/' => [
        'template' => 'about.latte',
        'title' => '{about.seo.title}',
        'description' => '{about.seo.description}',
        'keywords' => '{about.seo.keywords}',
        'data' => ['site', 'navigation', 'about'],
    ],
    '/projects' => [
        'template' => 'projects.latte',
        'title' => '{projects.seo.title}',
        'description' => '{projects.seo.description}',
        'keywords' => '{projects.seo.keywords}',
        'data' => ['site', 'navigation', 'projects'],
    ],
    '/projects/' => [
        'template' => 'projects.latte',
        'title' => '{projects.seo.title}',
        'description' => '{projects.seo.description}',
        'keywords' => '{projects.seo.keywords}',
        'data' => ['site', 'navigation', 'projects'],
    ],
    '/contact' => [
        'template' => 'contact.latte',
        'title' => '{contact.seo.title}',
        'description' => '{contact.seo.description}',
        'keywords' => '{contact.seo.keywords}',
        'data' => ['site', 'navigation', 'contact', 'about'],
    ],
    '/contact/' => [
        'template' => 'contact.latte',
        'title' => '{contact.seo.title}',
        'description' => '{contact.seo.description}',
        'keywords' => '{contact.seo.keywords}',
        'data' => ['site', 'navigation', 'contact', 'about'],
    ],
];
