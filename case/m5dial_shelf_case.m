% M5Stack Dial under-shelf case — MATLAB drawing
% Units: millimetres. Run in MATLAB: m5dial_shelf_case
% Produces a 4-view mechanical drawing plus an isometric.

function m5dial_shelf_case
    wall = 3; top_t = 5; depth = 2*25.4; usb_clear = 15;
    dial_bezel_d = 51; dial_hole_d = 45.2;
    screw_d = 3; access_d = 10; cable_d = 8;
    ped_t = 3; ped_hole_d = 8.2; ped_splay = 15;
    ped_span = 14; ped_w = 34; ped_h = 18;
    bottom_clear = 3; width = 86; screw_x = 33; screw_y = 18;
    bezel_r = dial_bezel_d/2;
    center_z = wall + bottom_clear + bezel_r;
    inner_top = center_z + bezel_r + usb_clear;
    height = inner_top + top_t;
    cable_z = inner_top - 3 - cable_d/2;

    fig = figure('Color','w','Name','M5Stack Dial shelf case','Position',[80 60 1400 900]);
    tiledlayout(fig,2,3,'Padding','compact','TileSpacing','compact');

    nexttile; draw_front(width,height,center_z,dial_hole_d,dial_bezel_d,screw_x,screw_d,wall,top_t);
    title('Front'); xlabel('X mm'); ylabel('Z mm'); axis equal tight; grid on; box on;

    nexttile; draw_top(width,depth,ped_t,screw_x,screw_y,screw_d,dial_bezel_d,dial_hole_d,ped_w,ped_span,ped_hole_d,ped_splay);
    title('Top'); xlabel('X mm'); ylabel('Y mm'); axis equal tight; grid on; box on;

    nexttile; draw_right(depth,height,ped_t,center_z,dial_hole_d,wall,top_t,usb_clear,bezel_r,cable_z,cable_d,screw_y,screw_d,ped_h);
    title('Right'); xlabel('Y mm'); ylabel('Z mm'); axis equal tight; grid on; box on;

    nexttile; draw_back(width,height,center_z,cable_z,cable_d,ped_w,ped_h,ped_span,ped_hole_d,ped_splay,wall,top_t,screw_x,screw_d);
    title('Back'); xlabel('X mm'); ylabel('Z mm'); axis equal tight; grid on; box on;

    nexttile([1 2]);
    draw_iso(width,depth,height,ped_t,ped_w,ped_h,center_z,dial_hole_d,screw_x,screw_y,screw_d,cable_z,cable_d,ped_span,ped_hole_d,ped_splay,wall,top_t,access_d);
    title(sprintf(['Isometric  |  %g x %g x %g mm  |  top %g mm  |  walls %g mm  |  ' ...
        'USB gap %g mm  |  screws 3 mm  |  cable 8 mm  |  pedestal 8.2 mm @ \\pm15\\circ'], ...
        width, depth, height, top_t, wall, usb_clear));
    xlabel('X mm'); ylabel('Y mm'); zlabel('Z mm'); axis equal vis3d; grid on; box on;
    view(38, 22); camlight('headlight'); lighting gouraud;

    sgtitle(fig, 'M5Stack Dial under-shelf case  (mm)');
    if ~batchStartupOptionUsed
        % save next to this script when run interactively
    end
    out = fullfile(fileparts(mfilename('fullpath')), 'm5dial_shelf_case_matlab.png');
    exportgraphics(fig, out, 'Resolution', 150);
    fprintf('Wrote %s\n', out);
end

function draw_front(width,height,cz,hole_d,bezel_d,sx,sd,wall,top_t)
    hold on;
    rectangle('Position',[-width/2 0 width height],'EdgeColor','k','LineWidth',1.4);
    rectangle('Position',[-width/2 height-top_t width top_t],'EdgeColor',[0.2 0.2 0.7],'LineStyle','--');
    viscircles([0 cz], hole_d/2, 'Color',[0.85 0.33 0.1], 'LineWidth',1.6);
    viscircles([0 cz], bezel_d/2, 'Color',[1 0.55 0.15], 'LineStyle','--', 'LineWidth',0.8);
    for s = [-1 1]
        viscircles([s*sx height], 0.01); % tick only
        plot([s*sx s*sx],[height-top_t height],'b','LineWidth',1.2);
        plot(s*sx, height-top_t/2, 'bo', 'MarkerSize', 4);
    end
    % dimensions
    dimh(0, -width/2, width/2, height+8, sprintf('%.1f', width));
    dimv(-width/2-8, 0, height, sprintf('%.1f', height));
    dimv(width/2+6, cz-hole_d/2, cz+hole_d/2, sprintf('\\oslash%.1f', hole_d));
    text(sx, height+4, '3 mm screws', 'HorizontalAlignment','center', 'FontSize',8, 'Color','b');
    hold off;
end

function draw_top(width,depth,ped_t,sx,sy,sd,bezel_d,hole_d,ped_w,span,phd,splay)
    hold on;
    rectangle('Position',[-width/2 0 width depth],'EdgeColor','k','LineWidth',1.4);
    rectangle('Position',[-ped_w/2 depth ped_w ped_t],'EdgeColor',[0.1 0.5 0.2],'LineWidth',1.2);
    viscircles([0 0], hole_d/2, 'Color',[0.85 0.33 0.1], 'LineWidth',1.0); % front hole edge-on: skip
    plot([-hole_d/2 hole_d/2],[0 0],'Color',[0.85 0.33 0.1],'LineWidth',2);
    for s = [-1 1]
        viscircles([s*sx sy], sd/2, 'Color','b', 'LineWidth',1.3);
        a = splay * s * pi/180;
        x0 = s*span/2; y0 = depth;
        x1 = x0 + 12*sin(a); y1 = y0 + ped_t*cos(a) + 6;
        plot([x0 x1],[y0 y1],'Color',[0.1 0.5 0.2],'LineWidth',1.4);
        viscircles([s*span/2 depth+ped_t/2], phd/2, 'Color',[0.1 0.5 0.2], 'LineWidth',1.0);
    end
    dimh(0, -width/2, width/2, -8, sprintf('%.1f', width));
    dimv(width/2+8, 0, depth, sprintf('%.1f  (2.00 in)', depth));
    text(0, sy+6, 'screws L/R of dial', 'HorizontalAlignment','center', 'FontSize',8, 'Color','b');
    text(0, depth+ped_t+7, '8.2 mm @ \pm15°', 'HorizontalAlignment','center', 'FontSize',8, 'Color',[0.1 0.5 0.2]);
    ylim([-12 depth+ped_t+14]);
    hold off;
end

function draw_right(depth,height,ped_t,cz,hole_d,wall,top_t,usb,bezel_r,cable_z,cable_d,sy,sd,ped_h)
    hold on;
    rectangle('Position',[0 0 depth height],'EdgeColor','k','LineWidth',1.4);
    rectangle('Position',[depth cz-ped_h/2 ped_t ped_h],'EdgeColor',[0.1 0.5 0.2],'LineWidth',1.2);
    rectangle('Position',[0 height-top_t depth top_t],'EdgeColor',[0.2 0.2 0.7],'LineStyle','--');
    % dial hole on front wall
    plot([0 wall],[cz-hole_d/2 cz-hole_d/2],'Color',[0.85 0.33 0.1],'LineWidth',1.4);
    plot([0 wall],[cz+hole_d/2 cz+hole_d/2],'Color',[0.85 0.33 0.1],'LineWidth',1.4);
    % USB clearance band
    y1 = cz + bezel_r; y2 = height-top_t;
    patch([wall depth-wall depth-wall wall],[y1 y1 y2 y2],[1 0.9 0.7],'EdgeColor','none','FaceAlpha',0.35);
    plot([depth-wall depth],[cable_z cable_z],'k','LineWidth',1.2);
    viscircles([depth cable_z], cable_d/2, 'Color','k', 'LineWidth',1.0);
    plot([sy sy],[height-top_t height],'b','LineWidth',1.2);
    dimv(-8, y1, y2, sprintf('%g USB', usb));
    dimv(depth+ped_t+8, 0, height, sprintf('%.1f', height));
    dimh(height+8, 0, depth, sprintf('%.1f', depth));
    text(depth/2, (y1+y2)/2, '15 mm USB', 'HorizontalAlignment','center', 'FontSize',8);
    hold off;
end

function draw_back(width,height,cz,cable_z,cable_d,ped_w,ped_h,span,phd,splay,wall,top_t,sx,sd)
    hold on;
    rectangle('Position',[-width/2 0 width height],'EdgeColor','k','LineWidth',1.4);
    rectangle('Position',[-ped_w/2 cz-ped_h/2 ped_w ped_h],'EdgeColor',[0.1 0.5 0.2],'LineWidth',1.4);
    viscircles([0 cable_z], cable_d/2, 'Color','k', 'LineWidth',1.5);
    text(0, cable_z+7, '\oslash8 mm cable', 'HorizontalAlignment','center', 'FontSize',8);
    for s = [-1 1]
        viscircles([s*span/2 cz], phd/2, 'Color',[0.1 0.5 0.2], 'LineWidth',1.4);
        plot(s*sx, height-top_t/2, 'bo', 'MarkerSize', 4);
    end
    text(0, cz-ped_h/2-6, '8.2 mm holes  30° included  (\pm15° from rear axis)', ...
        'HorizontalAlignment','center', 'FontSize',8, 'Color',[0.1 0.5 0.2]);
    hold off;
end

function draw_iso(width,depth,height,ped_t,ped_w,ped_h,cz,hole_d,sx,sy,sd,cable_z,cable_d,span,phd,splay,wall,top_t,access_d)
    hold on;
    % outer box faces (MATLAB default blue)
    c = [0 0.447 0.741];
    x = [-width/2 width/2]; y = [0 depth]; z = [0 height];
    % six faces
    patch('XData',[x(1) x(2) x(2) x(1)],'YData',[y(1) y(1) y(1) y(1)],'ZData',[z(1) z(1) z(2) z(2)], ...
        'FaceColor',c,'FaceAlpha',0.55,'EdgeColor','k'); % front
    patch('XData',[x(1) x(2) x(2) x(1)],'YData',[y(2) y(2) y(2) y(2)],'ZData',[z(1) z(1) z(2) z(2)], ...
        'FaceColor',c,'FaceAlpha',0.45,'EdgeColor','k'); % back
    patch('XData',[x(1) x(1) x(1) x(1)],'YData',[y(1) y(2) y(2) y(1)],'ZData',[z(1) z(1) z(2) z(2)], ...
        'FaceColor',c,'FaceAlpha',0.35,'EdgeColor','k');
    patch('XData',[x(2) x(2) x(2) x(2)],'YData',[y(1) y(2) y(2) y(1)],'ZData',[z(1) z(1) z(2) z(2)], ...
        'FaceColor',c,'FaceAlpha',0.50,'EdgeColor','k');
    patch('XData',[x(1) x(2) x(2) x(1)],'YData',[y(1) y(1) y(2) y(2)],'ZData',[z(2) z(2) z(2) z(2)], ...
        'FaceColor',[0.3 0.3 0.35],'FaceAlpha',0.55,'EdgeColor','k'); % top
    patch('XData',[x(1) x(2) x(2) x(1)],'YData',[y(1) y(1) y(2) y(2)],'ZData',[z(1) z(1) z(1) z(1)], ...
        'FaceColor',c,'FaceAlpha',0.25,'EdgeColor','k');
    % pedestal
    patch('XData',[-ped_w/2 ped_w/2 ped_w/2 -ped_w/2], ...
          'YData',[depth+ped_t depth+ped_t depth+ped_t depth+ped_t], ...
          'ZData',[cz-ped_h/2 cz-ped_h/2 cz+ped_h/2 cz+ped_h/2], ...
          'FaceColor',[0.47 0.67 0.19],'FaceAlpha',0.7,'EdgeColor','k');
    patch('XData',[-ped_w/2 ped_w/2 ped_w/2 -ped_w/2], ...
          'YData',[depth depth depth+ped_t depth+ped_t], ...
          'ZData',[cz+ped_h/2 cz+ped_h/2 cz+ped_h/2 cz+ped_h/2], ...
          'FaceColor',[0.47 0.67 0.19],'FaceAlpha',0.7,'EdgeColor','k');
    % front hole
    circ3(0,0,cz,hole_d/2,[0 1 0], [0.85 0.33 0.1]);
    % top screws
    for s = [-1 1]
        circ3(s*sx, sy, height, sd/2, [0 0 1], [0 0.3 0.9]);
        circ3(s*sx, sy, 0, access_d/2, [0 0 1], [0.2 0.2 0.2]);
        % pedestal holes, angled
        a = s*splay*pi/180;
        circ3(s*span/2, depth+ped_t, cz, phd/2, [sin(a) cos(a) 0], [0.1 0.5 0.2]);
    end
    circ3(0, depth, cable_z, cable_d/2, [0 1 0], [0.1 0.1 0.1]);
    hold off;
end

function circ3(x,y,z,r,normal,col)
    n = normal(:) / norm(normal);
    if abs(n(1)) < 0.9
        a = cross(n, [1 0 0]');
    else
        a = cross(n, [0 1 0]');
    end
    a = a / norm(a); b = cross(n, a);
    t = linspace(0, 2*pi, 48);
    P = [x;y;z] + r * (a * cos(t) + b * sin(t));
    plot3(P(1,:), P(2,:), P(3,:), 'Color', col, 'LineWidth', 1.6);
end

function dimh(~, x1, x2, y, label)
    plot([x1 x2],[y y],'k-');
    plot(x1,y,'k<','MarkerFaceColor','k','MarkerSize',4);
    plot(x2,y,'k>','MarkerFaceColor','k','MarkerSize',4);
    text((x1+x2)/2, y+0.5, label, 'HorizontalAlignment','center', 'FontSize',8, 'BackgroundColor','w');
end

function dimv(x, y1, y2, label)
    plot([x x],[y1 y2],'k-');
    plot(x,y1,'kv','MarkerFaceColor','k','MarkerSize',4);
    plot(x,y2,'k^','MarkerFaceColor','k','MarkerSize',4);
    text(x+0.5, (y1+y2)/2, label, 'HorizontalAlignment','left', 'FontSize',8, 'BackgroundColor','w');
end
