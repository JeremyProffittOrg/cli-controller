% Adafruit PCA9548 8-channel STEMMA QT / Qwiic I2C multiplexer (5626) case
% MATLAB mechanical drawing. Units: millimetres. Run in MATLAB: pca9548_case
%
% Open-top tray in two heights, 12.00 mm and 24.00 mm, with a 3.00 mm buffer
% around the board on every side.
%
% Eight 3 mm x 3 mm cable exits are notched into the TOP of the two long walls,
% from max(z) - 3 up to the rim, on the eight JST SH channel-port centres.
% The two short ends carry mounting tabs instead of cutouts: each reaches
% 10 mm out, is 3 mm high in the same max(z) - 3 band, has a 4 mm hole midway,
% and is rounded at its outer end.
%
% Board geometry from adafruit/Adafruit-PCA9548-PCB, PCA9548A QT Board.brd
% (Eagle 9.6.2): layer-20 outline, the four mounting holes, the nine JST_SH4
% elements. Mirrors pca9548_case.scad and draw_pca9548_case.py.

function pca9548_case
    P = case_params();
    fig = figure('Color','w','Name','PCA9548 I2C mux case','Position',[50 30 1620 1000]);
    tiledlayout(fig,2,3,'Padding','compact','TileSpacing','compact');

    nexttile; draw_plan(P);
    title(sprintf('Plan  (X-Y)  open top, %.2f mm gap around the board', P.gap));
    xlabel('X mm'); ylabel('Y mm'); axis equal tight; grid on; box on;

    nexttile; draw_long(P, P.h_std);
    title(sprintf('Long side  (X-Z)  std, %.2f mm', P.h_std));
    xlabel('X mm'); ylabel('Z mm'); axis equal tight; grid on; box on;

    nexttile; draw_long(P, P.h_tall);
    title(sprintf('Long side  (X-Z)  tall, %.2f mm', P.h_tall));
    xlabel('X mm'); ylabel('Z mm'); axis equal tight; grid on; box on;

    nexttile; draw_end(P, P.h_std);
    title(sprintf('End  (Y-Z)  tab at max(z)-%.0f, no cable exit', P.tab_h));
    xlabel('Y mm'); ylabel('Z mm'); axis equal tight; grid on; box on;

    nexttile([1 2]); draw_iso(P, P.h_tall);
    title(sprintf('Isometric  |  body %.2f x %.2f, %.2f over tabs, h %.2f mm', ...
        P.outer_x, P.outer_y, P.outer_x + 2*P.tab_out, P.h_tall));
    xlabel('X mm'); ylabel('Y mm'); zlabel('Z mm');
    axis equal vis3d; grid on; box on;
    view(38, 24); camlight('headlight'); lighting gouraud;

    sgtitle(fig, sprintf(['Adafruit PCA9548 (5626) 8-channel I2C mux case  |  ' ...
        'body %.2f x %.2f mm, %.2f over tabs  |  h %.0f and h %.0f mm  |  ' ...
        '8 x 3x3 mm exits notched from the rim'], ...
        P.outer_x, P.outer_y, P.outer_x + 2*P.tab_out, P.h_std, P.h_tall));

    out = fullfile(fileparts(mfilename('fullpath')), 'pca9548_case_matlab.png');
    exportgraphics(fig, out, 'Resolution', 150);
    fprintf('Wrote %s\n', out);
end

function P = case_params()
    % Adafruit PCA9548 breakout, from the Eagle board file
    P.pcb_x = 40.64; P.pcb_y = 20.32; P.pcb_t = 1.60; P.pcb_r = 2.54;
    P.hole_span_x = 35.56; P.hole_span_y = 15.24; P.pcb_hole_d = 2.032;
    P.port_x = [-11.43 -3.81 3.81 11.43];   % JST SH4 channel centres, centred
    P.jst_w = 6.20; P.jst_h = 2.90;

    % case
    P.wall = 3.00; P.floor_t = 2.00; P.fit = 0.50; P.buffer = 3.00;
    P.h_std = 12.00; P.h_tall = 24.00;
    P.cut = 3.00; P.cut_r = 0.60;
    P.stand_h = 3.00; P.stand_od = 5.00;
    P.stand_hole = 2.00; P.stand_hole_depth = 4.00;

    % end mounting tabs
    P.tab_out = 10.00; P.tab_h = 3.00; P.tab_r = 5.00;
    P.tab_hole_d = 4.00; P.tab_hole_out = 5.00;

    P.gap = P.fit + P.buffer;               % 3.50 mm board edge to inner wall
    P.inner_x = P.pcb_x + 2*P.gap;
    P.inner_y = P.pcb_y + 2*P.gap;
    P.outer_x = P.inner_x + 2*P.wall;
    P.outer_y = P.inner_y + 2*P.wall;
    P.inner_r = P.pcb_r + P.gap;
    P.outer_r = P.inner_r + P.wall;

    P.pcb_z0 = P.floor_t + P.stand_h;       % board underside, 5.00
    P.pcb_z1 = P.pcb_z0 + P.pcb_t;          % board top face, 6.60

    P.blue   = [0 0.447 0.741];
    P.orange = [0.85 0.325 0.098];
    P.green  = [0.466 0.674 0.188];
    P.purple = [0.494 0.184 0.556];
    P.grey   = [0.4 0.4 0.42];
end

function z = cut_z0(P, height)
    z = height - P.cut;                     % exits notch down from the rim
end

function z = tab_z0(P, height)
    z = height - P.tab_h;
end

function pts = tab_outline(P, sx)
    root = sx * P.outer_x/2;
    centre = root + sx * (P.tab_out - P.tab_r);
    t = linspace(-pi/2, pi/2, 48);
    ax_ = centre + sx * P.tab_r * cos(t);
    ay_ = P.tab_r * sin(t);
    pts = [ [root, ax_, root]', [-P.tab_r, ay_, P.tab_r]' ];
end

function rrect(x, y, w, h, r, colour, lw, style)
    rectangle('Position',[x y w h], ...
        'Curvature',[2*r/w 2*r/h], ...
        'EdgeColor',colour,'LineWidth',lw,'LineStyle',style);
end

function draw_plan(P)
    hold on;
    ox = P.outer_x/2; oy = P.outer_y/2;

    for sx = [-1 1]
        pts = tab_outline(P, sx);
        patch('XData',pts(:,1),'YData',pts(:,2), ...
            'FaceColor',[0.94 0.90 0.97],'EdgeColor',P.purple,'LineWidth',1.5);
        viscircles([sx*(ox + P.tab_hole_out) 0], P.tab_hole_d/2, ...
            'Color',P.purple,'LineWidth',1.4);
    end

    rrect(-ox, -oy, P.outer_x, P.outer_y, P.outer_r, 'k', 1.4, '-');
    rrect(-P.inner_x/2, -P.inner_y/2, P.inner_x, P.inner_y, P.inner_r, P.grey, 1.0, '--');
    rrect(-P.pcb_x/2, -P.pcb_y/2, P.pcb_x, P.pcb_y, P.pcb_r, P.green, 1.6, '-');

    for sx = [-1 1]
        for sy = [-1 1]
            cx = sx*P.hole_span_x/2; cy = sy*P.hole_span_y/2;
            viscircles([cx cy], P.stand_od/2, 'Color', P.blue, 'LineWidth', 1.3);
            viscircles([cx cy], P.stand_hole/2, 'Color', P.blue, 'LineWidth', 1.1);
        end
    end

    for sy = [-1 1]
        for k = 1:numel(P.port_x)
            px = P.port_x(k);
            y0 = min(sy*P.inner_y/2, sy*oy);
            rectangle('Position',[px-P.cut/2 y0 P.cut P.wall], ...
                'FaceColor',[1 0.90 0.84],'EdgeColor',P.orange,'LineWidth',1.3);
        end
    end

    dimh(-ox, ox, oy+8.0, sprintf('%.2f', P.outer_x), 'k');
    dimh(-ox-P.tab_out, ox+P.tab_out, oy+12.5, ...
        sprintf('%.2f over tabs', P.outer_x + 2*P.tab_out), P.purple);
    dimh(ox, ox+P.tab_out, -oy-3.5, sprintf('%.2f', P.tab_out), P.purple);
    dimh(-P.pcb_x/2, P.pcb_x/2, -oy-8.0, sprintf('PCB %.2f', P.pcb_x), P.green);
    dimv(-ox-P.tab_out-4.0, -oy, oy, sprintf('%.2f', P.outer_y), 'k');
    dimh(P.port_x(1), P.port_x(2), oy+1.6, '7.62', P.orange);
    dimh(-P.hole_span_x/2, P.hole_span_x/2, 3.4, sprintf('%.2f', P.hole_span_x), P.blue);
    dimv(-6.5, -P.hole_span_y/2, P.hole_span_y/2, sprintf('%.2f', P.hole_span_y), P.blue);
    text(-P.pcb_x/2-P.gap/2, P.pcb_y/2+P.gap/2, sprintf('%.2f gap', P.gap), ...
        'HorizontalAlignment','center','FontSize',7.5,'Color',P.grey,'BackgroundColor','w');
    hold off;
end

function draw_long(P, height)
    hold on;
    ox = P.outer_x/2;
    cz0 = cut_z0(P, height);
    tz0 = tab_z0(P, height);

    for sx = [-1 1]
        root = sx*ox;
        rectangle('Position',[min(root, root+sx*P.tab_out) tz0 P.tab_out P.tab_h], ...
            'FaceColor',[0.94 0.90 0.97],'EdgeColor',P.purple,'LineWidth',1.5);
        hx = sx*(ox + P.tab_hole_out);
        for dx = [-P.tab_hole_d/2 P.tab_hole_d/2]
            plot([hx+dx hx+dx],[tz0 tz0+P.tab_h],':','Color',P.purple,'LineWidth',1.0);
        end
    end

    rectangle('Position',[-ox 0 P.outer_x height], ...
        'FaceColor',[0.94 0.96 0.99],'EdgeColor','k','LineWidth',1.4);
    plot([-ox ox],[P.floor_t P.floor_t],'--','Color',P.grey,'LineWidth',1.0);

    % the notches open onto the rim, so draw three sides only
    for k = 1:numel(P.port_x)
        px = P.port_x(k);
        rectangle('Position',[px-P.cut/2 cz0 P.cut P.cut+1.0], ...
            'FaceColor','w','EdgeColor','none');
        plot([px-P.cut/2 px-P.cut/2 px+P.cut/2 px+P.cut/2], ...
             [height cz0+P.cut_r cz0+P.cut_r height], ...
             '-','Color',P.orange,'LineWidth',1.5);
        rectangle('Position',[px-P.jst_w/2 P.pcb_z0-P.jst_h P.jst_w P.jst_h], ...
            'EdgeColor',P.green,'LineStyle',':','LineWidth',0.8);
    end

    rectangle('Position',[-P.pcb_x/2 P.pcb_z0 P.pcb_x P.pcb_t], ...
        'FaceColor',[0.90 0.96 0.88],'EdgeColor',P.green,'LineStyle','--','LineWidth',1.2);

    for sx = [-1 1]
        cx = sx*P.hole_span_x/2;
        rectangle('Position',[cx-P.stand_od/2 P.floor_t P.stand_od P.stand_h], ...
            'FaceColor',[0.88 0.93 0.98],'EdgeColor',P.blue,'LineWidth',1.1);
        plot([cx cx],[P.pcb_z0-P.stand_hole_depth P.pcb_z0],'-','Color',P.blue,'LineWidth',1.6);
    end

    dimv(ox+P.tab_out+3.5, 0, height, sprintf('%.2f', height), 'k');
    dimv(-ox-2.6, cz0, height, sprintf('%.2f', P.cut), P.orange);
    dimv(-ox-P.tab_out-2.6, 0, P.floor_t, sprintf('%.2f', P.floor_t), P.grey);
    text(0, height+1.4, sprintf('exits Z %.2f to %.2f', cz0, height), ...
        'HorizontalAlignment','center','FontSize',8,'Color',P.orange);
    text(0, P.pcb_z1+0.8, sprintf('board top face Z=%.2f', P.pcb_z1), ...
        'HorizontalAlignment','center','FontSize',8,'Color',P.green);
    hold off;
end

function draw_end(P, height)
    hold on;
    oy = P.outer_y/2;
    tz0 = tab_z0(P, height);
    rectangle('Position',[-oy 0 P.outer_y height], ...
        'FaceColor',[0.94 0.96 0.99],'EdgeColor','k','LineWidth',1.4);
    plot([-oy oy],[P.floor_t P.floor_t],'--','Color',P.grey,'LineWidth',1.0);
    rectangle('Position',[-P.tab_r tz0 2*P.tab_r P.tab_h], ...
        'FaceColor',[0.94 0.90 0.97],'EdgeColor',P.purple,'LineWidth',1.5);
    rectangle('Position',[-P.tab_hole_d/2 tz0 P.tab_hole_d P.tab_h], ...
        'FaceColor','w','EdgeColor',P.purple,'LineStyle',':','LineWidth',1.0);
    rectangle('Position',[-P.pcb_y/2 P.pcb_z0 P.pcb_y P.pcb_t], ...
        'FaceColor',[0.90 0.96 0.88],'EdgeColor',P.green,'LineStyle','--','LineWidth',1.2);
    for sy = [-1 1]
        cy = sy*P.hole_span_y/2;
        rectangle('Position',[cy-P.stand_od/2 P.floor_t P.stand_od P.stand_h], ...
            'FaceColor',[0.88 0.93 0.98],'EdgeColor',P.blue,'LineWidth',1.1);
    end
    dimh(-oy, oy, height+2.6, sprintf('%.2f', P.outer_y), 'k');
    dimh(-P.tab_r, P.tab_r, tz0-1.8, sprintf('tab %.2f wide', 2*P.tab_r), P.purple);
    dimv(oy+3.0, 0, height, sprintf('%.2f', height), 'k');
    dimv(-oy-3.0, tz0, height, sprintf('%.2f', P.tab_h), P.purple);
    hold off;
end

function draw_iso(P, height)
    hold on;
    ox = P.outer_x/2; oy = P.outer_y/2;
    ix = P.inner_x/2; iy = P.inner_y/2;
    cz0 = cut_z0(P, height);
    tz0 = tab_z0(P, height);
    box3(-ox, ox, -oy, oy, 0, P.floor_t, P.blue, 0.30);
    box3(-ox, ox, iy, oy, P.floor_t, height, P.blue, 0.18);
    box3(-ox, ox, -oy, -iy, P.floor_t, height, P.blue, 0.18);
    box3(-ox, -ix, -iy, iy, P.floor_t, height, P.blue, 0.18);
    box3( ix, ox, -iy, iy, P.floor_t, height, P.blue, 0.18);
    for k = 1:numel(P.port_x)
        px = P.port_x(k);
        for sy = [-1 1]
            lo = min(sy*iy, sy*oy); hi = max(sy*iy, sy*oy);
            box3(px-P.cut/2, px+P.cut/2, lo, hi, cz0, height, P.orange, 0.85);
        end
    end
    for sx = [-1 1]
        prism(tab_outline(P, sx), tz0, height, P.purple, 0.55);
    end
    for sx = [-1 1]
        for sy = [-1 1]
            cx = sx*P.hole_span_x/2; cy = sy*P.hole_span_y/2;
            box3(cx-P.stand_od/2, cx+P.stand_od/2, cy-P.stand_od/2, cy+P.stand_od/2, ...
                P.floor_t, P.pcb_z0, P.blue, 0.55);
        end
    end
    box3(-P.pcb_x/2, P.pcb_x/2, -P.pcb_y/2, P.pcb_y/2, P.pcb_z0, P.pcb_z1, P.green, 0.45);
    hold off;
end

function box3(x0, x1, y0, y1, z0, z1, colour, alpha)
    V = [x0 y0 z0; x1 y0 z0; x1 y1 z0; x0 y1 z0; ...
         x0 y0 z1; x1 y0 z1; x1 y1 z1; x0 y1 z1];
    F = [1 2 3 4; 5 6 7 8; 1 2 6 5; 4 3 7 8; 1 4 8 5; 2 3 7 6];
    patch('Vertices',V,'Faces',F,'FaceColor',colour,'FaceAlpha',alpha, ...
        'EdgeColor',[0.1 0.1 0.1],'LineWidth',0.5);
end

function prism(pts, z0, z1, colour, alpha)
    n = size(pts,1);
    V = [pts, repmat(z0,n,1); pts, repmat(z1,n,1)];
    F = cell(n+1,1);
    F{1} = 1:n;
    F{2} = (n+1):(2*n);
    for i = 1:(n-1)
        F{i+2} = [i i+1 n+i+1 n+i];
    end
    for i = 1:numel(F)
        patch('Vertices',V,'Faces',F{i},'FaceColor',colour,'FaceAlpha',alpha, ...
            'EdgeColor',[0.1 0.1 0.1],'LineWidth',0.4);
    end
end

function dimh(x1, x2, y, label, colour)
    plot([x1 x2],[y y],'-','Color',colour);
    plot(x1,y,'<','Color',colour,'MarkerFaceColor',colour,'MarkerSize',4);
    plot(x2,y,'>','Color',colour,'MarkerFaceColor',colour,'MarkerSize',4);
    text((x1+x2)/2, y, label, 'HorizontalAlignment','center', ...
        'FontSize',8, 'Color',colour, 'BackgroundColor','w');
end

function dimv(x, y1, y2, label, colour)
    plot([x x],[y1 y2],'-','Color',colour);
    plot(x,y1,'v','Color',colour,'MarkerFaceColor',colour,'MarkerSize',4);
    plot(x,y2,'^','Color',colour,'MarkerFaceColor',colour,'MarkerSize',4);
    text(x, (y1+y2)/2, label, 'HorizontalAlignment','center','Rotation',90, ...
        'FontSize',8, 'Color',colour, 'BackgroundColor','w');
end
