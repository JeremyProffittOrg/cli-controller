% Adafruit PCA9548 8-channel STEMMA QT / Qwiic I2C multiplexer (5626) case
% MATLAB mechanical drawing. Units: millimetres. Run in MATLAB: pca9548_case
%
% Open-top tray in two heights, 12.00 mm and 24.00 mm. Board drops in
% component-side down onto four standoffs. Nine 3 mm x 3 mm cable exits sit on
% the floor at Z 2.00 to 5.00 in BOTH heights: four in each long wall on the
% eight JST SH port centres, one in the left end wall on the controller port.
%
% Board geometry from adafruit/Adafruit-PCA9548-PCB, PCA9548A QT Board.brd
% (Eagle 9.6.2): layer-20 outline, the four mounting holes, the nine JST_SH4
% elements. Mirrors pca9548_case.scad and draw_pca9548_case.py.

function pca9548_case
    P = case_params();
    fig = figure('Color','w','Name','PCA9548 I2C mux case','Position',[60 40 1560 980]);
    tiledlayout(fig,2,3,'Padding','compact','TileSpacing','compact');

    nexttile; draw_plan(P);
    title('Plan  (X-Y)  open top, board component-side down');
    xlabel('X mm'); ylabel('Y mm'); axis equal tight; grid on; box on;

    nexttile; draw_long(P, P.h_std);
    title(sprintf('Long side  (X-Z)  std, %.2f mm', P.h_std));
    xlabel('X mm'); ylabel('Z mm'); axis equal tight; grid on; box on;

    nexttile; draw_long(P, P.h_tall);
    title(sprintf('Long side  (X-Z)  tall, %.2f mm', P.h_tall));
    xlabel('X mm'); ylabel('Z mm'); axis equal tight; grid on; box on;

    nexttile; draw_end(P);
    title('Left end  (Y-Z)  controller-side port exit, on centre');
    xlabel('Y mm'); ylabel('Z mm'); axis equal tight; grid on; box on;

    nexttile([1 2]); draw_iso(P, P.h_tall);
    title(sprintf('Isometric  |  %.2f x %.2f x %.2f mm, tall variant', ...
        P.outer_x, P.outer_y, P.h_tall));
    xlabel('X mm'); ylabel('Y mm'); zlabel('Z mm');
    axis equal vis3d; grid on; box on;
    view(38, 24); camlight('headlight'); lighting gouraud;

    sgtitle(fig, sprintf(['Adafruit PCA9548 (5626) 8-channel I2C mux case  |  ' ...
        '%.2f x %.2f mm  |  h %.0f and h %.0f mm  |  9 x 3x3 mm exits at Z %.0f-%.0f'], ...
        P.outer_x, P.outer_y, P.h_std, P.h_tall, P.cut_z0, P.cut_z1));

    out = fullfile(fileparts(mfilename('fullpath')), 'pca9548_case_matlab.png');
    exportgraphics(fig, out, 'Resolution', 150);
    fprintf('Wrote %s\n', out);
end

function P = case_params()
    % Adafruit PCA9548 breakout, from the Eagle board file
    P.pcb_x = 40.64; P.pcb_y = 20.32; P.pcb_t = 1.60; P.pcb_r = 2.54;
    P.hole_span_x = 35.56; P.hole_span_y = 15.24; P.pcb_hole_d = 2.032;
    P.port_x = [-11.43 -3.81 3.81 11.43];   % JST SH4 centres, centred model
    P.jst_w = 6.20; P.jst_h = 2.90;

    % case
    P.wall = 3.00; P.floor_t = 2.00; P.fit = 0.50;
    P.h_std = 12.00; P.h_tall = 24.00;
    P.cut = 3.00; P.cut_r = 0.60;
    P.stand_h = 3.00; P.stand_od = 5.00;
    P.stand_hole = 2.00; P.stand_hole_depth = 4.00;

    P.inner_x = P.pcb_x + 2*P.fit;
    P.inner_y = P.pcb_y + 2*P.fit;
    P.outer_x = P.inner_x + 2*P.wall;
    P.outer_y = P.inner_y + 2*P.wall;
    P.inner_r = P.pcb_r + P.fit;
    P.outer_r = P.inner_r + P.wall;

    P.pcb_z0 = P.floor_t + P.stand_h;   % board underside, 5.00
    P.pcb_z1 = P.pcb_z0 + P.pcb_t;      % board top face, 6.60
    P.cut_z0 = P.floor_t;               % 2.00
    P.cut_z1 = P.cut_z0 + P.cut;        % 5.00, flush with the board underside

    P.blue   = [0 0.447 0.741];
    P.orange = [0.85 0.325 0.098];
    P.green  = [0.466 0.674 0.188];
    P.grey   = [0.4 0.4 0.42];
end

function rrect(x, y, w, h, r, colour, lw, style)
    rectangle('Position',[x y w h], ...
        'Curvature',[2*r/w 2*r/h], ...
        'EdgeColor',colour,'LineWidth',lw,'LineStyle',style);
end

function draw_plan(P)
    hold on;
    ox = P.outer_x/2; oy = P.outer_y/2;
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
    rectangle('Position',[-ox -P.cut/2 P.wall P.cut], ...
        'FaceColor',[1 0.90 0.84],'EdgeColor',P.orange,'LineWidth',1.3);

    dimh(-ox, ox, oy+5.0, sprintf('%.2f', P.outer_x), 'k');
    dimh(-P.pcb_x/2, P.pcb_x/2, -oy-4.0, sprintf('PCB %.2f', P.pcb_x), P.green);
    dimv(-ox-5.0, -oy, oy, sprintf('%.2f', P.outer_y), 'k');
    dimh(P.port_x(1), P.port_x(2), oy+1.4, '7.62', P.orange);
    dimh(-P.hole_span_x/2, P.hole_span_x/2, 3.2, sprintf('%.2f', P.hole_span_x), P.blue);
    dimv(-6.5, -P.hole_span_y/2, P.hole_span_y/2, sprintf('%.2f', P.hole_span_y), P.blue);
    hold off;
end

function draw_long(P, height)
    hold on;
    ox = P.outer_x/2;
    rectangle('Position',[-ox 0 P.outer_x height], ...
        'FaceColor',[0.94 0.96 0.99],'EdgeColor','k','LineWidth',1.4);
    plot([-ox ox],[P.floor_t P.floor_t],'--','Color',P.grey,'LineWidth',1.0);

    for k = 1:numel(P.port_x)
        px = P.port_x(k);
        rectangle('Position',[px-P.cut/2 P.cut_z0 P.cut P.cut], ...
            'Curvature',[2*P.cut_r/P.cut 2*P.cut_r/P.cut], ...
            'FaceColor','w','EdgeColor',P.orange,'LineWidth',1.5);
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

    dimv(ox+3.0, 0, height, sprintf('%.2f', height), 'k');
    dimv(-ox-3.0, P.cut_z0, P.cut_z1, sprintf('%.2f', P.cut), P.orange);
    dimv(-ox-7.5, 0, P.floor_t, sprintf('%.2f', P.floor_t), P.grey);
    text(0, height+1.2, sprintf('board top face Z=%.2f', P.pcb_z1), ...
        'HorizontalAlignment','center','FontSize',8,'Color',P.green);
    hold off;
end

function draw_end(P)
    hold on;
    oy = P.outer_y/2;
    rectangle('Position',[-oy 0 P.outer_y P.h_std], ...
        'FaceColor',[0.94 0.96 0.99],'EdgeColor','k','LineWidth',1.4);
    plot([-oy oy],[P.floor_t P.floor_t],'--','Color',P.grey,'LineWidth',1.0);
    rectangle('Position',[-P.cut/2 P.cut_z0 P.cut P.cut], ...
        'Curvature',[2*P.cut_r/P.cut 2*P.cut_r/P.cut], ...
        'FaceColor','w','EdgeColor',P.orange,'LineWidth',1.5);
    rectangle('Position',[-P.pcb_y/2 P.pcb_z0 P.pcb_y P.pcb_t], ...
        'FaceColor',[0.90 0.96 0.88],'EdgeColor',P.green,'LineStyle','--','LineWidth',1.2);
    for sy = [-1 1]
        cy = sy*P.hole_span_y/2;
        rectangle('Position',[cy-P.stand_od/2 P.floor_t P.stand_od P.stand_h], ...
            'FaceColor',[0.88 0.93 0.98],'EdgeColor',P.blue,'LineWidth',1.1);
    end
    dimh(-oy, oy, P.h_std+2.2, sprintf('%.2f', P.outer_y), 'k');
    dimv(oy+3.0, 0, P.h_std, sprintf('%.2f', P.h_std), 'k');
    dimv(-P.cut/2-2.2, P.cut_z0, P.cut_z1, sprintf('%.2f', P.cut), P.orange);
    hold off;
end

function draw_iso(P, height)
    hold on;
    ox = P.outer_x/2; oy = P.outer_y/2;
    ix = P.inner_x/2; iy = P.inner_y/2;
    box3(-ox, ox, -oy, oy, 0, P.floor_t, P.blue, 0.30);
    box3(-ox, ox, iy, oy, P.floor_t, height, P.blue, 0.18);
    box3(-ox, ox, -oy, -iy, P.floor_t, height, P.blue, 0.18);
    box3(-ox, -ix, -iy, iy, P.floor_t, height, P.blue, 0.18);
    box3( ix, ox, -iy, iy, P.floor_t, height, P.blue, 0.18);
    for k = 1:numel(P.port_x)
        px = P.port_x(k);
        for sy = [-1 1]
            lo = min(sy*iy, sy*oy); hi = max(sy*iy, sy*oy);
            box3(px-P.cut/2, px+P.cut/2, lo, hi, P.cut_z0, P.cut_z1, P.orange, 0.85);
        end
    end
    box3(-ox, -ix, -P.cut/2, P.cut/2, P.cut_z0, P.cut_z1, P.orange, 0.85);
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
